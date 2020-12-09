package lite

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	hub "github.com/sentinel-official/hub/types"
	"github.com/sentinel-official/hub/x/node"
	"github.com/sentinel-official/hub/x/plan"
	"github.com/sentinel-official/hub/x/subscription"
	"github.com/sentinel-official/hub/x/vpn"
)

func (c *Client) QueryAccount(address sdk.AccAddress) (auth.Account, error) {
	bytes, err := c.ctx.Codec.MarshalJSON(auth.NewQueryAccountParams(address))
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("custom/%s/%s", auth.QuerierRoute, auth.QueryAccount)
	res, _, err := c.ctx.QueryWithData(path, bytes)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}

	var item auth.Account
	if err := c.ctx.Codec.UnmarshalJSON(res, &item); err != nil {
		return nil, err
	}

	return item, nil
}

func (c *Client) QueryNode(address hub.NodeAddress) (*node.Node, error) {
	bytes, err := c.ctx.Codec.MarshalJSON(node.NewQueryNodeParams(address))
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("custom/%s/%s/%s", vpn.StoreKey, node.QuerierRoute, node.QueryNode)
	res, _, err := c.ctx.QueryWithData(path, bytes)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}

	var item node.Node
	if err := c.ctx.Codec.UnmarshalJSON(res, &item); err != nil {
		return nil, err
	}

	return &item, nil
}

func (c *Client) QuerySubscription(id uint64) (*subscription.Subscription, error) {
	bytes, err := c.ctx.Codec.MarshalJSON(subscription.NewQuerySubscriptionParams(id))
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("custom/%s/%s/%s", vpn.StoreKey, subscription.QuerierRoute, subscription.QuerySubscription)
	res, _, err := c.ctx.QueryWithData(path, bytes)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}

	var item subscription.Subscription
	if err := c.ctx.Codec.UnmarshalJSON(res, &item); err != nil {
		return nil, err
	}

	return &item, nil
}

func (c *Client) QueryQuota(id uint64, address sdk.AccAddress) (*subscription.Quota, error) {
	bytes, err := c.ctx.Codec.MarshalJSON(subscription.NewQueryQuotaParams(id, address))
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("custom/%s/%s/%s", vpn.StoreKey, subscription.QuerierRoute, subscription.QueryQuota)
	res, _, err := c.ctx.QueryWithData(path, bytes)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}

	var item subscription.Quota
	if err := c.ctx.Codec.UnmarshalJSON(res, &item); err != nil {
		return nil, err
	}

	return &item, nil
}

func (c *Client) HasNodeForPlan(id uint64, address hub.NodeAddress) (bool, error) {
	res, _, err := c.ctx.QueryStore(plan.NodeForPlanKey(id, address), vpn.ModuleName)
	if err != nil {
		return false, err
	}
	if res == nil {
		return false, nil
	}

	var item bool
	if err := c.ctx.Codec.UnmarshalJSON(res, &item); err != nil {
		return false, err
	}

	return item, nil
}
