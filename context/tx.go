package context

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	hubtypes "github.com/sentinel-official/hub/types"
	nodetypes "github.com/sentinel-official/hub/x/node/types"
	sessiontypes "github.com/sentinel-official/hub/x/session/types"

	"github.com/sentinel-official/dvpn-node/types"
)

func (c *Context) RegisterNode() error {
	c.Log().Info("Registering the node...")

	_, err := c.Client().Tx(
		nodetypes.NewMsgRegisterRequest(
			c.Operator(),
			c.GigabytePrices(),
			c.HourlyPrices(),
			c.RemoteURL(),
		),
	)
	if err != nil {
		c.Log().Error("failed to register the node", "error", err)
		return err
	}

	return nil
}

func (c *Context) UpdateNodeInfo() error {
	c.Log().Info("Updating the node info...")

	_, err := c.Client().Tx(
		nodetypes.NewMsgUpdateDetailsRequest(
			c.Address(),
			c.GigabytePrices(),
			c.HourlyPrices(),
			c.RemoteURL(),
		),
	)
	if err != nil {
		c.Log().Error("failed to update the node info", "error", err)
		return err
	}

	return nil
}

func (c *Context) UpdateNodeStatus() error {
	c.Log().Info("Updating the node status...")

	_, err := c.Client().Tx(
		nodetypes.NewMsgUpdateStatusRequest(
			c.Address(),
			hubtypes.StatusActive,
		),
	)
	if err != nil {
		c.Log().Error("failed to update the node status", "error", err)
		return err
	}

	return nil
}

func (c *Context) UpdateSessions(items ...types.Session) error {
	c.Log().Info("Updating the sessions...")

	messages := make([]sdk.Msg, 0, len(items))
	for _, item := range items {
		messages = append(messages,
			sessiontypes.NewMsgUpdateDetailsRequest(
				c.Address(),
				sessiontypes.Proof{
					ID:        item.ID,
					Duration:  item.UpdatedAt.Sub(item.CreatedAt),
					Bandwidth: hubtypes.NewBandwidthFromInt64(item.Upload, item.Download),
				},
				nil,
			),
		)
	}

	_, err := c.Client().Tx(
		messages...,
	)
	if err != nil {
		c.Log().Error("failed to update the sessions", "error", err)
		return err
	}

	return nil
}
