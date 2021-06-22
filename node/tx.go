package node

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	hubtypes "github.com/sentinel-official/hub/types"
	nodetypes "github.com/sentinel-official/hub/x/node/types"
	sessiontypes "github.com/sentinel-official/hub/x/session/types"

	"github.com/sentinel-official/dvpn-node/types"
)

func (n *Node) register() error {
	n.Log().Info("Registering node...")

	_, err := n.Client().BroadcastTx(
		nodetypes.NewMsgRegisterRequest(
			n.Operator(),
			n.Provider(),
			n.Price(),
			n.RemoteURL(),
		),
	)

	return err
}

func (n *Node) updateInfo() error {
	n.Log().Info("Updating node info...")

	_, err := n.Client().BroadcastTx(
		nodetypes.NewMsgUpdateRequest(
			n.Address(),
			n.Provider(),
			n.Price(),
			n.RemoteURL(),
		),
	)

	return err
}

func (n *Node) updateStatus() error {
	n.Log().Info("Updating node status...")

	_, err := n.Client().BroadcastTx(
		nodetypes.NewMsgSetStatusRequest(
			n.Address(),
			hubtypes.StatusActive,
		),
	)

	return err
}

func (n *Node) updateSessions(items ...types.Session) error {
	n.Log().Info("Updating sessions...")

	var messages []sdk.Msg
	for _, item := range items {
		messages = append(messages,
			sessiontypes.NewMsgUpdateRequest(
				n.Address(),
				sessiontypes.Proof{
					Id:        item.ID,
					Duration:  time.Since(item.ConnectedAt),
					Bandwidth: hubtypes.NewBandwidthFromInt64(item.Download, item.Upload),
				},
				nil,
			),
		)
	}

	_, err := n.Client().BroadcastTx(messages...)
	return err
}
