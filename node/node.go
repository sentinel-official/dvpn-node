package node

import (
	"fmt"
	"log"
	"net/http"
	"time"

	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
	vpnTypes "github.com/ironman0x7b2/sentinel-sdk/x/vpn/types"
	"github.com/tendermint/tendermint/libs/common"

	"github.com/ironman0x7b2/vpn-node/tx"
	"github.com/ironman0x7b2/vpn-node/types"
)

type Node struct {
	id      sdkTypes.ID
	owner   csdkTypes.AccAddress
	apiPort uint16

	tx       *tx.Tx
	vpn      types.BaseVPN
	sessions types.Sessions
}

func NewNode(details *vpnTypes.NodeDetails, tx *tx.Tx, vpn types.BaseVPN) *Node {
	return &Node{
		id:      details.ID,
		owner:   details.Owner,
		apiPort: details.APIPort,

		tx:       tx,
		vpn:      vpn,
		sessions: types.NewSessions(),
	}
}

func (n Node) Start() {
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
		if err := n.updateSessionsBandwidth(); err != nil {
			panic(err)
		}
	}()

	addr := fmt.Sprintf("0.0.0.0:%d", n.apiPort)

	log.Printf("Listening the API server on address `%s`", addr)
	if err := http.ListenAndServe(addr, n.Router()); err != nil {
		panic(err)
	}
}

func (n Node) updateNodeStatus() error {
	log.Printf("Starting update node status ticker with interval %s", types.IntervalUpdateNodeStatus.String())
	t := time.NewTicker(types.IntervalUpdateNodeStatus)

	for ; ; <-t.C {
		msg := vpnTypes.NewMsgUpdateNodeStatus(n.owner, n.id, vpnTypes.StatusActive)
		data, err := n.tx.CompleteAndSubscribeTx(msg)
		if err != nil {
			return err
		}

		log.Printf("Node status info updated at block height `%d`, tx hash `%s`",
			data.Height, common.HexBytes(data.Tx.Hash()).String())
	}
}

func (n Node) updateSessionsBandwidth() error {
	log.Printf("Starting update sessions bandwidth ticker with interval %s", types.IntervalUpdateSessionsBandwidth.String())
	t1 := time.NewTicker(types.IntervalUpdateSessionsBandwidth)

	log.Printf("Starting request bandwidth signs ticker with interval %s", types.IntervalRequestBandwidthSigns.String())
	t2 := time.NewTicker(types.IntervalRequestBandwidthSigns)

	var makeTx bool
	for {
		<-t2.C

		clients, err := n.vpn.ClientList()
		if err != nil {
			return err
		}

		ids := n.sessions.IDs()
		msgs := make([]csdkTypes.Msg, 0, len(ids))

		select {
		case <-t1.C:
			makeTx = true
		default:
			makeTx = false
		}

		for _, id := range ids {
			session := n.sessions.Get(id)
			if session == nil || session.Status == vpnTypes.StatusInactive {
				n.sessions.Delete(id)
			}
			if session.Status == vpnTypes.StatusInit {
				continue
			}
			if session.Status == vpnTypes.StatusActive {
				go func() {
					if err := n.requestBandwidthSign(session, clients[id]); err != nil {
						panic(err)
					}
				}()
			}

			if makeTx {
				bandwidth, nodeOwnerSign, clientSign := session.BandwidthInfo()
				msg := vpnTypes.NewMsgUpdateSessionBandwidth(session.NodeOwner, session.ID,
					bandwidth.Upload, bandwidth.Download, nodeOwnerSign, clientSign)
				msgs = append(msgs, msg)
			}
		}

		if makeTx && len(msgs) > 0 {
			go func() {
				data, err := n.tx.CompleteAndSubscribeTx(msgs...)
				if err != nil {
					panic(err)
				}

				log.Printf("Sessions bandwidth info updated at block height `%d`, tx hash `%s`",
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

	session.OutMessages <- NewMsgBandwidthSign(session.ID.String(), bandwidth, sign, "").GetBytes()
	return nil
}
