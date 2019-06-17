package node

import (
	"fmt"
	"log"
	"net/http"
	"time"

	csdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/common"

	sdk "github.com/ironman0x7b2/sentinel-sdk/types"
	"github.com/ironman0x7b2/sentinel-sdk/x/vpn"

	_db "github.com/ironman0x7b2/vpn-node/db"
	_tx "github.com/ironman0x7b2/vpn-node/tx"
	"github.com/ironman0x7b2/vpn-node/types"
)

type Node struct {
	id      sdk.ID
	address csdk.AccAddress
	pubKey  crypto.PubKey

	tx      *_tx.Tx
	db      *_db.DB
	vpn     types.BaseVPN
	clients map[string]*client
}

func NewNode(id sdk.ID, address csdk.AccAddress, pubKey crypto.PubKey,
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

func (n *Node) Start(apiPort uint16) error {
	if err := n.vpn.Init(); err != nil {
		return err
	}

	go func() {
		if err := n.vpn.Start(); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := n.updateNodeStatus(); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := n.updateAllSessionBandwidthsInfo(); err != nil {
			panic(err)
		}
	}()

	addr := fmt.Sprintf("0.0.0.0:%d", apiPort)

	log.Printf("Listening the API server on address `%s`", addr)
	return http.ListenAndServe(addr, n.Router())
}

func (n *Node) updateNodeStatus() error {
	log.Printf("Starting update node status ticker with interval `%s`",
		types.UpdateNodeStatusInterval.String())

	t := time.NewTicker(types.UpdateNodeStatusInterval)
	for ; ; <-t.C {
		msg := vpn.NewMsgUpdateNodeStatus(n.address, n.id, vpn.StatusActive)

		data, err := n.tx.CompleteAndSubscribeTx(msg)
		if err != nil {
			return err
		}

		log.Printf("Node status updated at block height `%d`, tx hash `%s`",
			data.Height, common.HexBytes(data.Tx.Hash()).String())
	}
}

func (n *Node) updateAllSessionBandwidthsInfo() error {
	// TODO: From VPN
	return nil
}

func (n *Node) requestBandwidthSign(session *types.Session, bandwidth sdk.Bandwidth) error {
	return nil
}
