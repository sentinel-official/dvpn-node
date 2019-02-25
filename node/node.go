package node

import (
	"fmt"
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
	*vpnTypes.NodeDetails
	tx       *tx.Tx
	vpn      types.BaseVPN
	sessions types.Sessions
	clients  types.Clients
}

func NewNode(details *vpnTypes.NodeDetails, tx *tx.Tx, vpn types.BaseVPN) *Node {
	return &Node{
		NodeDetails: details,
		tx:          tx,
		vpn:         vpn,
		sessions:    types.NewSessions(),
		clients:     types.NewClients(),
	}
}

func (n Node) Start() {
	go func() {
		if err := n.updateNodeStatus(); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := n.requestBandwidthSigns(); err != nil {
			panic(err)
		}
	}()

	if err := http.ListenAndServe(":8000", n.Router()); err != nil {
		panic(err)
	}
}

func (n Node) updateNodeStatus() error {
	t := time.NewTicker(types.IntervalUpdateNodeStatus)

	for range t.C {
		msg := vpnTypes.NewMsgUpdateNodeStatus(n.Owner, n.ID, vpnTypes.StatusActive)
		data, err := n.tx.CompleteAndSubscribeTx(msg)
		if err != nil {
			return err
		}

		fmt.Println(data.Height, common.HexBytes(data.Tx.Hash()).String())
	}

	return nil
}

func (n Node) updateSessionsBandwidth(clients []types.VPNClient) error {
	if len(clients) == 0 {
		return nil
	}

	msgs := make([]csdkTypes.Msg, len(clients))
	for _, c := range clients {
		session := n.sessions.Get(c.ID)
		if session == nil {
			continue
		}

		bandwidth, nodeOwnerSign, clientSign := session.BandwidthInfo()
		msg := vpnTypes.NewMsgUpdateSessionBandwidth(session.NodeOwner, session.ID,
			bandwidth.Upload, bandwidth.Download, nodeOwnerSign, clientSign)
		msgs = append(msgs, msg)
	}

	_, err := n.tx.CompleteAndSubscribeTx(msgs...)

	return err
}

func (n Node) requestBandwidthSigns() error {
	t1 := time.NewTicker(types.IntervalRequestBandwidthSigns)
	t2 := time.NewTicker(types.IntervalUpdateSessionsBandwidth)

	for {
		clients, err := n.vpn.ClientList()
		if err != nil {
			return err
		}

		select {
		case <-t1.C:
			for index := range clients {
				go func(client *types.VPNClient) {
					if err := n.requestBandwidthSign(client); err != nil {
						panic(err)
					}
				}(&clients[index])
			}
		case <-t2.C:
			go func() {
				if err := n.updateSessionsBandwidth(clients); err != nil {
					panic(err)
				}
			}()
		}
	}
}

func (n Node) requestBandwidthSign(c *types.VPNClient) error {
	session := n.sessions.Get(c.ID)
	if session == nil {
		_ = n.vpn.DisconnectClient(c.ID)
		return nil
	}

	client := n.clients.Get(c.ID)
	if client == nil {
		return nil
	}

	sign, err := n.tx.SignSessionBandwidth(session.ID, c.Upload, c.Download, session.Client)
	if err != nil {
		return err
	}

	bandwidth := sdkTypes.NewBandwidthFromInt64(c.Upload, c.Download)
	client.OutMessages <- NewMsgBandwidthSign(c.ID, bandwidth, sign, "").GetBytes()

	return nil
}
