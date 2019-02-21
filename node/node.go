package node

import (
	"fmt"
	"net/http"
	"time"

	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	vpnTypes "github.com/ironman0x7b2/sentinel-sdk/x/vpn/types"

	"github.com/ironman0x7b2/vpn-node/tx"
	"github.com/ironman0x7b2/vpn-node/types"
)

type Node struct {
	*vpnTypes.NodeDetails
	tx       *tx.Tx
	vpn      types.BaseVPN
	sessions map[string]*types.Session
	clients  map[string]*types.Client
}

func NewNode(details *vpnTypes.NodeDetails, tx *tx.Tx, vpn types.BaseVPN) *Node {
	return &Node{
		NodeDetails: details,
		tx:          tx,
		vpn:         vpn,
		sessions:    make(map[string]*types.Session),
		clients:     make(map[string]*types.Client),
	}
}

func (n Node) Start() {
	errors := make(chan error)

	go n.jobUpdateNodeStatus(errors)
	go n.jobSendSessionBandwidthInfo(errors)
	go n.jobUpdateSessionsBandwidth(errors)

	go func() {
		panic(<-errors)
	}()

	if err := http.ListenAndServe(":8000", n.Router()); err != nil {
		panic(err)
	}
}

func (n Node) jobUpdateNodeStatus(errors chan error) {
	for ; ; time.Sleep(200 * time.Second) {
		msg := vpnTypes.NewMsgUpdateNodeStatus(n.Owner, n.ID, vpnTypes.StatusActive)
		_, err := n.tx.CompleteAndSubscribeTx([]csdkTypes.Msg{msg})
		if err != nil {
			errors <- err
			return
		}
	}
}

func (n Node) jobUpdateSessionsBandwidth(errors chan error) {
	for ; ; time.Sleep(100 * time.Second) {
		vpnClients, err := n.vpn.ClientList()
		if err != nil {
			errors <- err
			return
		}

		var msgs []csdkTypes.Msg

		for _, vc := range vpnClients {
			session := n.sessions[vc.ID]
			if session == nil {
				_ = n.vpn.DisconnectClient(vc.ID)
				continue
			}

			bandwidth, nodeOwnerSign, clientSign := session.BandwidthSigns()

			msg := vpnTypes.NewMsgUpdateSessionBandwidth(session.NodeOwner, session.ID,
				bandwidth.Upload, bandwidth.Download, nodeOwnerSign, clientSign)
			msgs = append(msgs, msg)
		}

		if len(msgs) == 0 {
			continue
		}

		data, err := n.tx.CompleteAndSubscribeTx(msgs)
		if err != nil {
			errors <- err
			return
		}

		fmt.Println(data.Result.IsOK())
	}
}

func (n Node) jobSendSessionBandwidthInfo(errors chan error) {
	for ; ; time.Sleep(5 * time.Second) {
		vpnClients, err := n.vpn.ClientList()
		if err != nil {
			errors <- err
			return
		}

		for _, vc := range vpnClients {
			client := n.clients[vc.ID]
			session := n.sessions[vc.ID]

			if client == nil {
				continue
			}

			sign, err := n.tx.SignSessionBandwidth(session.ID, vc.Upload, vc.Download, session.Client)
			if err != nil {
				errors <- err
				return
			}

			client.OutMessages <- NewMsgBandwidthSign(vc.ID, vc.Upload, vc.Download, sign, "").Bytes()
		}
	}
}
