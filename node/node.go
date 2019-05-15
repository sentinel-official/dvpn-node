package node

import (
	"fmt"
	"log"
	"net/http"
	"time"

	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
	"github.com/ironman0x7b2/sentinel-sdk/x/vpn"
	"github.com/tendermint/tendermint/libs/common"

	"github.com/ironman0x7b2/vpn-node/tx"
	"github.com/ironman0x7b2/vpn-node/types"
)

type Node struct {
	id       sdkTypes.ID
	owner    csdkTypes.AccAddress
	tx       *tx.Tx
	vpn      types.BaseVPN
	sessions types.Sessions
}

func NewNode(node *vpn.Node, tx *tx.Tx, vpn types.BaseVPN) *Node {
	return &Node{
		id:       node.ID,
		owner:    node.Owner,
		tx:       tx,
		vpn:      vpn,
		sessions: types.NewSessions(),
	}
}

func (n Node) Start(apiPort uint16) {
	if err := n.vpn.Init(); err != nil {
		panic(err)
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

	listenAddress := fmt.Sprintf("0.0.0.0:%d", apiPort)

	log.Printf("Listening the API server on address `%s`", listenAddress)
	if err := http.ListenAndServe(listenAddress, n.Router()); err != nil {
		panic(err)
	}
}

func (n Node) updateNodeStatus() error {
	log.Printf("Starting update node status ticker with interval `%s`",
		types.UpdateNodeStatusInterval.String())

	t := time.NewTicker(types.UpdateNodeStatusInterval)
	for ; ; <-t.C {
		msg := vpn.NewMsgUpdateNodeStatus(n.owner, n.id, vpn.StatusActive)

		data, err := n.tx.CompleteAndSubscribeTx(msg)
		if err != nil {
			return err
		}

		log.Printf("Node status updated at block height `%d`, tx hash `%s`",
			data.Height, common.HexBytes(data.Tx.Hash()).String())
	}
}

func (n Node) updateAllSessionBandwidthsInfo() error {
	log.Printf("Starting update all session bandwidths info ticker with interval `%s`",
		types.UpdateSessionBandwidthInfoInterval.String())
	t1 := time.NewTicker(types.UpdateSessionBandwidthInfoInterval)

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

		clients, err := n.vpn.ClientList()
		if err != nil {
			return err
		}

		ids := n.sessions.IDs()
		messages := make([]csdkTypes.Msg, 0, len(ids))

		for _, id := range ids {
			session := n.sessions.Get(id)
			if session == nil || session.Status == vpn.StatusInactive {
				n.sessions.Delete(id)
			}

			if session.Status == vpn.StatusInit {
				continue
			}
			if session.Status == vpn.StatusActive {
				go func() {
					if err := n.requestBandwidthSign(session, clients[id]); err != nil {
						panic(err)
					}
				}()
			}

			if makeTx {
				consumed, nodeOwnerSign, clientSign := session.ConsumedBandwidthInfo()
				message := vpn.NewMsgUpdateSessionBandwidthInfo(n.owner, session.ID, consumed, nodeOwnerSign, clientSign)
				messages = append(messages, message)
			}
		}

		if makeTx && len(messages) > 0 {
			go func() {
				data, err := n.tx.CompleteAndSubscribeTx(messages...)
				if err != nil {
					panic(err)
				}

				log.Printf("All Session bandwidths info updated at block height `%d`, tx hash `%s`",
					data.Height, common.HexBytes(data.Tx.Hash()).String())
			}()
		}
	}
}

func (n Node) requestBandwidthSign(session *types.Session, bandwidth sdkTypes.Bandwidth) error {
	sign, err := n.tx.SignSessionBandwidth(session.ID, bandwidth, session.Client)
	if err != nil {
		return err
	}

	session.OutMessages <- NewMsgBandwidthSign(session.ID.String(), bandwidth, sign, "")
	return nil
}
