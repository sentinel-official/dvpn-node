package node

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/common"

	hub "github.com/sentinel-official/hub/types"
	"github.com/sentinel-official/hub/x/vpn"

	_db "github.com/sentinel-official/dvpn-node/db"
	_tx "github.com/sentinel-official/dvpn-node/tx"
	"github.com/sentinel-official/dvpn-node/types"
)

type Node struct {
	id      hub.ID
	address sdk.AccAddress
	pubKey  crypto.PubKey

	tx      *_tx.Tx
	db      *_db.DB
	vpn     types.BaseVPN
	clients map[string]*client
}

func NewNode(id hub.ID, address sdk.AccAddress, pubKey crypto.PubKey,
	tx *_tx.Tx, db *_db.DB, _vpn types.BaseVPN) *Node {

	return &Node{
		id:      id,
		address: address,
		pubKey:  pubKey,

		tx:      tx,
		db:      db,
		vpn:     _vpn,
		clients: make(map[string]*client),
	}
}

func (n *Node) Start(port uint16) error {
	if err := n.vpn.Init(); err != nil {
		return err
	}

	go func() {
		if err := n.vpn.Start(); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := n.updateBandwidthInfos(); err != nil {
			panic(err)
		}
	}()

	addr := fmt.Sprintf("0.0.0.0:%d", port)

	log.Printf("Listening the API server on address `%s`", addr)
	return http.ListenAndServeTLS(addr, types.DefaultTLSCertFilePath, types.DefaultTLSKeyFilePath, n.Router())
}

func (n *Node) updateBandwidthInfos() error {
	log.Printf("Starting update bandwidth infos ticker with interval `%s`",
		types.UpdateBandwidthInfosInterval.String())
	t1 := time.NewTicker(types.UpdateBandwidthInfosInterval)

	log.Printf("Starting request bandwidth sign ticker with interval `%s`",
		types.RequestBandwidthSignInterval.String())
	t2 := time.NewTicker(types.RequestBandwidthSignInterval)

	var makeTx bool
	var wg sync.WaitGroup

	for ; ; <-t2.C {
		select {
		case <-t1.C:
			makeTx = true
		default:
			makeTx = false
		}

		clients, err := n.vpn.ClientsList()
		if err != nil {
			return err
		}

		var messages []sdk.Msg
		for id, bandwidth := range clients {
			wg.Add(1)

			go func(id string, bandwidth hub.Bandwidth, makeTx bool) {
				message, err := n.requestBandwidthSign(id, bandwidth, makeTx)
				if err != nil {
					panic(err)
				}

				messages = append(messages, message)
				wg.Done()
			}(id, bandwidth, makeTx)
		}

		wg.Wait()
		if makeTx && len(messages) > 0 {
			go func() {
				log.Println("Broadcasting update-session-bandwidth transaction", messages)
				data, err := n.tx.CompleteAndSubscribeTx(messages...)
				if err != nil {
					panic(err)
				}

				log.Printf("Bandwidth infos updated at block height `%d`, tx hash `%s`",
					data.Height, common.HexBytes(data.Tx.Hash()).String())
			}()
		}
	}
}

func (n *Node) requestBandwidthSign(id string, bandwidth hub.Bandwidth, makeTx bool) (msg sdk.Msg, err error) {
	query, args := "_id = ? AND _status = ?", []interface{}{
		id,
		types.ACTIVE,
	}

	s, err := n.db.SessionFindOne(query, args...)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, n.vpn.DisconnectClient(id)
	}

	client, ok := n.clients[id]
	if !ok {
		return nil, errors.Errorf("Client with id `%s` exists in database but not in memory", id)
	}

	_id := hub.NewSubscriptionID(s.ID.Uint64())
	if makeTx {
		signature, err := n.tx.SignSessionBandwidth(_id, s.Index, s.Bandwidth) // nolint:govet
		if err != nil {
			return nil, err
		}

		nos := auth.StdSignature{
			PubKey:    n.pubKey,
			Signature: signature,
		}
		cs := auth.StdSignature{
			PubKey:    client.pubKey,
			Signature: s.Signature,
		}

		msg = vpn.NewMsgUpdateSessionInfo(n.address, _id, s.Bandwidth, nos, cs)
	}

	updates := map[string]interface{}{
		"_upload":   bandwidth.Upload.Int64(),
		"_download": bandwidth.Download.Int64(),
	}

	if err = n.db.SessionFindOneAndUpdate(updates, query, args...); err != nil { // nolint:gocritic
		return nil, err
	}

	signature, err := n.tx.SignSessionBandwidth(_id, s.Index, bandwidth)
	if err != nil {
		return nil, err
	}

	client.outMessages <- NewMsgBandwidthSignature(s.ID, s.Index, s.Bandwidth, signature, nil)
	return msg, nil
}

type HealthResponse struct {
	Status string `json:"status"`
}
