package lite

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/sentinel-official/dvpn-node/utils"
	hubtypes "github.com/sentinel-official/hub/types"
	nodetypes "github.com/sentinel-official/hub/x/node/types"
	plantypes "github.com/sentinel-official/hub/x/plan/types"
	sessiontypes "github.com/sentinel-official/hub/x/session/types"
	subscriptiontypes "github.com/sentinel-official/hub/x/subscription/types"
	vpntypes "github.com/sentinel-official/hub/x/vpn/types"
)

func (c *Client) QueryAccount(address sdk.AccAddress) (authtypes.AccountI, error) {
	var (
		account authtypes.AccountI
		qc      = authtypes.NewQueryClient(c.ctx)
	)

	c.Log().Info("Querying account", "address", address)
	res, err := qc.Account(
		context.Background(),
		&authtypes.QueryAccountRequest{Address: address.String()},
	)
	if err != nil {
		return nil, utils.ValidError(err)
	}

	if err := c.ctx.InterfaceRegistry.UnpackAny(res.Account, &account); err != nil {
		return nil, err
	}

	return account, nil
}

func (c *Client) QueryNode(address hubtypes.NodeAddress) (*nodetypes.Node, error) {
	qc := nodetypes.NewQueryServiceClient(c.ctx)

	c.Log().Info("Querying node", "address", address)
	res, err := qc.QueryNode(
		context.Background(),
		nodetypes.NewQueryNodeRequest(address),
	)
	if err != nil {
		return nil, utils.ValidError(err)
	}

	return &res.Node, nil
}

func (c *Client) QuerySubscription(id uint64) (*subscriptiontypes.Subscription, error) {
	qc := subscriptiontypes.NewQueryServiceClient(c.ctx)

	c.Log().Info("Querying subscription", "id", id)
	res, err := qc.QuerySubscription(
		context.Background(),
		subscriptiontypes.NewQuerySubscriptionRequest(id),
	)
	if err != nil {
		return nil, utils.ValidError(err)
	}

	return &res.Subscription, nil
}

func (c *Client) QueryQuota(id uint64, address sdk.AccAddress) (*subscriptiontypes.Quota, error) {
	qc := subscriptiontypes.NewQueryServiceClient(c.ctx)

	c.Log().Info("Querying quota", "id", id, "address", address)
	res, err := qc.QueryQuota(
		context.Background(),
		subscriptiontypes.NewQueryQuotaRequest(id, address),
	)
	if err != nil {
		return nil, utils.ValidError(err)
	}

	return &res.Quota, nil
}

func (c *Client) QuerySession(id uint64) (*sessiontypes.Session, error) {
	qc := sessiontypes.NewQueryServiceClient(c.ctx)

	c.Log().Info("Querying session", "id", id)
	res, err := qc.QuerySession(
		context.Background(),
		sessiontypes.NewQuerySessionRequest(id),
	)
	if err != nil {
		return nil, utils.ValidError(err)
	}

	return &res.Session, nil
}

func (c *Client) HasNodeForPlan(id uint64, address hubtypes.NodeAddress) (bool, error) {
	value, _, err := c.ctx.QueryStore(
		append(
			[]byte(plantypes.ModuleName+"/"),
			plantypes.NodeForPlanKey(id, address)...,
		),
		vpntypes.ModuleName,
	)
	if err != nil {
		return false, err
	}
	if value == nil {
		return false, nil
	}

	return true, nil
}
