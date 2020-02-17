package node

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
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
		var ids []string
		var wg sync.WaitGroup
		for id, bandwidth := range clients {
			wg.Add(1)

			subs, err := n.tx.QuerySubscription(id)
			if err != nil {
				panic(err)
			}

			if !bandwidth.AllLTE(subs.RemainingBandwidth) {
				ids = append(ids, id)
			}

			go func(id string, bandwidth hub.Bandwidth, makeTx bool) {
				message, err := n.requestBandwidthSign(id, bandwidth, makeTx)
				if err != nil {
					panic(err)
				}
				if message != nil {
					messages = append(messages, message)
				}

				wg.Done()
			}(id, bandwidth, makeTx)
		}

		wg.Wait()
		if makeTx && len(messages) > 0 {
			go func() {
				data, err := n.tx.CompleteAndSubscribeTx(messages...)
				if err != nil {
					log.Println(err)
				}

				for _, id := range ids {
					if n.clients[id] != nil && n.clients[id].conn != nil {
						n.clients[id].conn.Close()
						delete(n.clients, id)
					}
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

	client := n.clients[id]
	_id := hub.NewSubscriptionID(s.ID.Uint64())

	if client != nil {
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

			_msg := vpn.NewMsgUpdateSessionInfo(n.address, _id, s.Bandwidth, nos, cs)
			if _msg.ValidateBasic() == nil {
				msg = _msg
			}
		}

		subs, err := n.tx.QuerySubscription(s.ID.String())
		if err != nil {
			return nil, err
		}

		if !bandwidth.AllLTE(subs.RemainingBandwidth) {
			bandwidth = subs.RemainingBandwidth
		}

		signature, err := n.tx.SignSessionBandwidth(_id, s.Index, bandwidth)
		if err != nil {
			return nil, err
		}

		if client.conn != nil {
			client.outMessages <- NewMsgBandwidthSignature(_id, s.Index, bandwidth, signature, nil)
		}
	}

	return msg, nil
}

type HealthResponse struct {
	Status string `json:"status"`
}
