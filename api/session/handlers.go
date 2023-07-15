package session

import (
	"fmt"
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	hubtypes "github.com/sentinel-official/hub/types"
	subscriptiontypes "github.com/sentinel-official/hub/x/subscription/types"

	"github.com/sentinel-official/dvpn-node/context"
	"github.com/sentinel-official/dvpn-node/types"
)

func HandlerAddSession(ctx *context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		if ctx.Service().PeerCount() >= ctx.Config().QOS.MaxPeers {
			err := fmt.Errorf("reached maximum peers limit %d", ctx.Config().QOS.MaxPeers)
			c.JSON(http.StatusBadRequest, types.NewResponseError(1, err))
			return
		}

		req, err := NewRequestAddSession(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, types.NewResponseError(2, err))
			return
		}

		item := types.Session{}
		ctx.Database().Model(
			&types.Session{},
		).Where(
			&types.Session{
				ID: req.URI.ID,
			},
		).First(&item)

		if item.ID != 0 {
			err = fmt.Errorf("peer for session %d already exist", req.URI.ID)
			c.JSON(http.StatusBadRequest, types.NewResponseError(3, err))
			return
		}

		item = types.Session{}
		ctx.Database().Model(
			&types.Session{},
		).Where(
			&types.Session{
				Key: req.Body.Key,
			},
		).First(&item)

		if item.ID != 0 {
			err = fmt.Errorf("peer %s for service already exist", req.Body.Key)
			c.JSON(http.StatusBadRequest, types.NewResponseError(3, err))
			return
		}

		account, err := ctx.Client().QueryAccount(req.AccAddress)
		if err != nil {
			c.JSON(http.StatusInternalServerError, types.NewResponseError(4, err))
			return
		}
		if account == nil {
			err = fmt.Errorf("account %s does not exist", req.AccAddress)
			c.JSON(http.StatusNotFound, types.NewResponseError(4, err))
			return
		}
		if account.GetPubKey() == nil {
			err = fmt.Errorf("public key for account %s does not exist", req.AccAddress)
			c.JSON(http.StatusNotFound, types.NewResponseError(4, err))
			return
		}
		if ok := account.GetPubKey().VerifySignature(sdk.Uint64ToBigEndian(req.URI.ID), req.Signature); !ok {
			err = fmt.Errorf("failed to verify the signature %s", req.Signature)
			c.JSON(http.StatusBadRequest, types.NewResponseError(4, err))
			return
		}

		session, err := ctx.Client().QuerySession(req.URI.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, types.NewResponseError(5, err))
			return
		}
		if session == nil {
			err = fmt.Errorf("session %d does not exist", req.URI.ID)
			c.JSON(http.StatusNotFound, types.NewResponseError(5, err))
			return
		}
		if !session.Status.Equal(hubtypes.StatusActive) {
			err = fmt.Errorf("invalid status for session %d; expected %s, got %s", session.ID, hubtypes.StatusActive, session.Status)
			c.JSON(http.StatusNotFound, types.NewResponseError(5, err))
			return
		}
		if session.Address != req.URI.AccAddress {
			err = fmt.Errorf("account address mismatch; expected %s, got %s", req.URI.AccAddress, session.Address)
			c.JSON(http.StatusBadRequest, types.NewResponseError(5, err))
			return
		}

		subscription, err := ctx.Client().QuerySubscription(session.SubscriptionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, types.NewResponseError(6, err))
			return
		}
		if subscription == nil {
			err = fmt.Errorf("subscription %d does not exist", session.SubscriptionID)
			c.JSON(http.StatusNotFound, types.NewResponseError(6, err))
			return
		}
		if !subscription.GetStatus().Equal(hubtypes.StatusActive) {
			err = fmt.Errorf("invalid status for subscription %d; expected %s, got %s", subscription.GetID(), hubtypes.StatusActive, subscription.GetStatus())
			c.JSON(http.StatusBadRequest, types.NewResponseError(6, err))
			return
		}

		if subscription.Type() == subscriptiontypes.TypeNode {
			subscription := subscription.(*subscriptiontypes.NodeSubscription)
			if subscription.NodeAddress != ctx.Address().String() {
				err = fmt.Errorf("node address mismatch; expected %s, got %s", ctx.Address(), subscription.NodeAddress)
				c.JSON(http.StatusBadRequest, types.NewResponseError(7, err))
				return
			}
		} else {
			subscription := subscription.(*subscriptiontypes.PlanSubscription)

			ok, err := ctx.Client().HasNodeForPlan(subscription.PlanID, ctx.Address())
			if err != nil {
				c.JSON(http.StatusInternalServerError, types.NewResponseError(7, err))
				return
			}
			if !ok {
				err = fmt.Errorf("node %s does not exist for plan %d", ctx.Address(), subscription.PlanID)
				c.JSON(http.StatusBadRequest, types.NewResponseError(7, err))
				return
			}
		}

		allocation, err := ctx.Client().QueryAllocation(subscription.GetID(), req.AccAddress)
		if err != nil {
			c.JSON(http.StatusInternalServerError, types.NewResponseError(8, err))
			return
		}
		if allocation == nil {
			err = fmt.Errorf("allocation for address %s does not exist", req.URI.AccAddress)
			c.JSON(http.StatusNotFound, types.NewResponseError(8, err))
			return
		}

		var items []types.Session
		ctx.Database().Model(
			&types.Session{},
		).Where(
			&types.Session{
				Subscription: subscription.GetID(),
				Address:      req.URI.AccAddress,
			},
		).Find(&items)

		for i := 0; i < len(items); i++ {
			if err = ctx.RemovePeerIfExists(items[i].Key); err != nil {
				c.JSON(http.StatusInternalServerError, types.NewResponseError(9, err))
				return
			}

			allocation.UtilisedBytes = allocation.UtilisedBytes.
				Add(sdk.NewInt(items[i].Download + items[i].Upload))
		}

		if allocation.UtilisedBytes.GTE(allocation.GrantedBytes) {
			err = fmt.Errorf("allocation exceeded; granted %s, utilised %s", allocation.GrantedBytes, allocation.UtilisedBytes)
			c.JSON(http.StatusBadRequest, types.NewResponseError(10, err))
			return
		}

		result, err := ctx.Service().AddPeer(req.Key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, types.NewResponseError(11, err))
			return
		}
		ctx.Log().Info("Added a new peer", "key", req.Body.Key, "count", ctx.Service().PeerCount())

		ctx.Database().Model(
			&types.Session{},
		).Create(
			&types.Session{
				ID:           req.URI.ID,
				Subscription: subscription.GetID(),
				Key:          req.Body.Key,
				Address:      req.URI.AccAddress,
				Available:    allocation.GrantedBytes.Sub(allocation.UtilisedBytes).Int64(),
			},
		)

		result = append(result, ctx.IPv4Address()...)
		result = append(result, ctx.Service().Info()...)
		c.JSON(http.StatusCreated, types.NewResponseResult(result))
	}
}
