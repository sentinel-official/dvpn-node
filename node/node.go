package node

import (
	"net/http"
	"path"

	sdk "github.com/cosmos/cosmos-sdk/types"
	hubtypes "github.com/sentinel-official/hub/types"
	nodetypes "github.com/sentinel-official/hub/x/node/types"
	sessiontypes "github.com/sentinel-official/hub/x/session/types"

	"github.com/sentinel-official/dvpn-node/context"
	"github.com/sentinel-official/dvpn-node/types"
)

type Node struct {
	*context.Context
}

func NewNode(ctx *context.Context) *Node {
	return &Node{ctx}
}

func (n *Node) Initialize() error {
	result, err := n.Client().QueryNode(n.Address())
	if err != nil {
		return err
	}
	if result == nil {
		return n.register()
	}

	return n.update()
}

func (n *Node) Start() error {
	go func() {
		if err := n.jobUpdateStatus(); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := n.jobUpdateSessions(); err != nil {
			panic(err)
		}
	}()

	var (
		certFile = path.Join(n.Home(), "tls.crt")
		keyFile  = path.Join(n.Home(), "tls.key")
	)

	n.Logger().Info("Started REST API server", "address", n.ListenOn())
	return http.ListenAndServeTLS(n.ListenOn(), certFile, keyFile, n.Router())
}

func (n *Node) register() error {
	res, err := n.Client().BroadcastTx(
		nodetypes.NewMsgRegisterRequest(
			n.Operator(),
			n.Provider(),
			n.Price(),
			n.RemoteURL(),
		),
	)
	if err != nil {
		return err
	}

	n.Logger().Info("Registered node", "tx_hash", res.TxHash)
	return nil
}

func (n *Node) update() error {
	res, err := n.Client().BroadcastTx(
		nodetypes.NewMsgUpdateRequest(
			n.Address(),
			n.Provider(),
			n.Price(),
			n.RemoteURL(),
		),
	)
	if err != nil {
		return err
	}

	n.Logger().Info("Updated node information", "tx_hash", res.TxHash)
	return nil
}

func (n *Node) updateStatus() error {
	res, err := n.Client().BroadcastTx(
		nodetypes.NewMsgSetStatusRequest(
			n.Address(),
			hubtypes.StatusActive,
		),
	)
	if err != nil {
		return err
	}

	n.Logger().Info("Updated node status", "tx_hash", res.TxHash)
	return nil
}

func (n *Node) updateSessions(items []types.Session) error {
	if len(items) == 0 {
		return nil
	}

	messages := make([]sdk.Msg, 0, len(items))
	for _, item := range items {
		messages = append(messages, sessiontypes.NewMsgUpsertRequest(
			sessiontypes.Proof{
				Channel:      0,
				Subscription: item.Subscription,
				Node:         n.Address().String(),
				Duration:     item.Duration,
				Bandwidth:    hubtypes.NewBandwidthFromInt64(item.Download, item.Upload),
			},
			item.Address,
			nil,
		))
	}

	res, err := n.Client().BroadcastTx(messages...)
	if err != nil {
		return err
	}

	n.Logger().Info("Updated sessions", "tx_hash", res.TxHash)
	return nil
}
