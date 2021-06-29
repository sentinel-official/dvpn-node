package context

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	hubtypes "github.com/sentinel-official/hub/types"
	nodetypes "github.com/sentinel-official/hub/x/node/types"
	sessiontypes "github.com/sentinel-official/hub/x/session/types"

	"github.com/sentinel-official/dvpn-node/types"
)

func (c *Context) RegisterNode() error {
	c.Log().Info("Registering node...")

	_, err := c.Client().Tx(
		nodetypes.NewMsgRegisterRequest(
			c.Operator(),
			c.Provider(),
			c.Price(),
			c.RemoteURL(),
		),
	)
	if err != nil {
		c.Log().Error("Failed to register node", "error", err)
		return err
	}

	return nil
}

func (c *Context) UpdateNodeInfo() error {
	c.Log().Info("Updating node info...")

	_, err := c.Client().Tx(
		nodetypes.NewMsgUpdateRequest(
			c.Address(),
			c.Provider(),
			c.Price(),
			c.RemoteURL(),
		),
	)
	if err != nil {
		c.Log().Error("Failed to update node info", "error", err)
		return err
	}

	return nil
}

func (c *Context) UpdateNodeStatus() error {
	c.Log().Info("Updating node status...")

	_, err := c.Client().Tx(
		nodetypes.NewMsgSetStatusRequest(
			c.Address(),
			hubtypes.StatusActive,
		),
	)
	if err != nil {
		c.Log().Error("Failed to update node status", "error", err)
		return err
	}

	return nil
}

func (c *Context) UpdateSessions(items ...types.Session) error {
	c.Log().Info("Updating sessions...")

	var messages []sdk.Msg
	for _, item := range items {
		messages = append(messages,
			sessiontypes.NewMsgUpdateRequest(
				c.Address(),
				sessiontypes.Proof{
					Id:        item.ID,
					Duration:  item.UpdatedAt.Sub(item.CreatedAt),
					Bandwidth: hubtypes.NewBandwidthFromInt64(item.Download, item.Upload),
				},
				nil,
			),
		)
	}

	_, err := c.Client().Tx(
		messages...,
	)
	if err != nil {
		c.Log().Error("Failed to update sessions", "error", err)
		return err
	}

	return nil
}
