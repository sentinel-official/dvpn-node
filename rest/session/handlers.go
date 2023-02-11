package session

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	hubtypes "github.com/sentinel-official/hub/types"
	"github.com/v2fly/v2ray-core/v5/common/uuid"

	"github.com/sentinel-official/dvpn-node/context"
	v2raytypes "github.com/sentinel-official/dvpn-node/services/v2ray/types"
	"github.com/sentinel-official/dvpn-node/types"
)

func HandlerAddSession(ctx *context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		if ctx.Service().PeersCount() >= ctx.Config().QOS.MaxPeers {
			err := fmt.Errorf("reached maximum peers limit; maximum %d", ctx.Config().QOS.MaxPeers)
			c.JSON(http.StatusBadRequest, types.NewResponseError(1, err))
			return
		}

		req, err := NewRequestAddSession(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, types.NewResponseError(2, err))
			return
		}

		key := base64.StdEncoding.EncodeToString(req.key)
		if ctx.Service().Type() == v2raytypes.Type {
			uid, err := uuid.ParseBytes(req.key[1:])
			if err != nil {
				c.JSON(http.StatusBadRequest, types.NewResponseError(2, err))
				return
			}

			key = uid.String()
		}

		item := types.Session{}
		ctx.Database().Model(
			&types.Session{},
		).Where(
			&types.Session{
				ID: req.id,
			},
		).First(&item)

		if item.ID != 0 {
			err = fmt.Errorf("peer for session %d already exist", req.id)
			c.JSON(http.StatusBadRequest, types.NewResponseError(3, err))
			return
		}

		item = types.Session{}
		ctx.Database().Model(
			&types.Session{},
		).Where(
			&types.Session{
				Key: key,
			},
		).First(&item)

		if item.ID != 0 {
			err = fmt.Errorf("key %s for service already exist", key)
			c.JSON(http.StatusBadRequest, types.NewResponseError(3, err))
			return
		}

		account, err := ctx.Client().QueryAccount(req.accAddress)
		if err != nil {
			c.JSON(http.StatusInternalServerError, types.NewResponseError(4, err))
			return
		}
		if account == nil {
			err = fmt.Errorf("account %s does not exist", req.accAddress)
			c.JSON(http.StatusNotFound, types.NewResponseError(4, err))
			return
		}
		if account.GetPubKey() == nil {
			err = fmt.Errorf("public key for account %s does not exist", req.accAddress)
			c.JSON(http.StatusNotFound, types.NewResponseError(4, err))
			return
		}
		if ok := account.GetPubKey().VerifySignature(sdk.Uint64ToBigEndian(req.id), req.signature); !ok {
			err = fmt.Errorf("failed to verify the signature %s", req.signature)
			c.JSON(http.StatusBadRequest, types.NewResponseError(4, err))
			return
		}

		session, err := ctx.Client().QuerySession(req.id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, types.NewResponseError(5, err))
			return
		}
		if session == nil {
			err = fmt.Errorf("session %d does not exist", req.id)
			c.JSON(http.StatusNotFound, types.NewResponseError(5, err))
			return
		}
		if !session.Status.Equal(hubtypes.StatusActive) {
			err = fmt.Errorf("invalid status for session %d; expected %s, got %s", session.Id, hubtypes.StatusActive, session.Status)
			c.JSON(http.StatusNotFound, types.NewResponseError(5, err))
			return
		}
		if session.Address != req.URI.AccAddress {
			err = fmt.Errorf("account address mismatch; expected %s, got %s", req.URI.AccAddress, session.Address)
			c.JSON(http.StatusBadRequest, types.NewResponseError(5, err))
			return
		}

		subscription, err := ctx.Client().QuerySubscription(session.Subscription)
		if err != nil {
			c.JSON(http.StatusInternalServerError, types.NewResponseError(6, err))
			return
		}
		if subscription == nil {
			err = fmt.Errorf("subscription %d does not exist", session.Subscription)
			c.JSON(http.StatusNotFound, types.NewResponseError(6, err))
			return
		}
		if !subscription.Status.Equal(hubtypes.StatusActive) {
			err = fmt.Errorf("invalid status for subscription %d; expected %s, got %s", subscription.Id, hubtypes.StatusActive, subscription.Status)
			c.JSON(http.StatusBadRequest, types.NewResponseError(6, err))
			return
		}

		if subscription.Plan == 0 {
			if subscription.Node != ctx.Address().String() {
				err = fmt.Errorf("node address mismatch; expected %s, got %s", ctx.Address(), subscription.Node)
				c.JSON(http.StatusBadRequest, types.NewResponseError(7, err))
				return
			}
		} else {
			ok, err := ctx.Client().HasNodeForPlan(subscription.Plan, ctx.Address())
			if err != nil {
				c.JSON(http.StatusInternalServerError, types.NewResponseError(7, err))
				return
			}
			if !ok {
				err = fmt.Errorf("node %s does not exist for plan %d", ctx.Address(), subscription.Plan)
				c.JSON(http.StatusBadRequest, types.NewResponseError(7, err))
				return
			}
		}

		quota, err := ctx.Client().QueryQuota(subscription.Id, req.accAddress)
		if err != nil {
			c.JSON(http.StatusInternalServerError, types.NewResponseError(8, err))
			return
		}
		if quota == nil {
			err = fmt.Errorf("quota for address %s does not exist", req.URI.AccAddress)
			c.JSON(http.StatusNotFound, types.NewResponseError(8, err))
			return
		}

		var items []types.Session
		ctx.Database().Model(
			&types.Session{},
		).Where(
			&types.Session{
				Subscription: subscription.Id,
				Address:      req.URI.AccAddress,
			},
		).Find(&items)

		for i := 0; i < len(items); i++ {
			session, err := ctx.Client().QuerySession(items[i].ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, types.NewResponseError(9, err))
				return
			}
			if session == nil {
				continue
			}

			if err = ctx.RemovePeer(items[i].Key); err != nil {
				c.JSON(http.StatusInternalServerError, types.NewResponseError(9, err))
				return
			}

			consumed := items[i].Download + items[i].Upload
			quota.Consumed = quota.Consumed.Add(
				hubtypes.NewBandwidthFromInt64(
					consumed, 0,
				).CeilTo(
					hubtypes.Gigabyte.Quo(subscription.Price.Amount),
				).Sum(),
			)
		}

		if quota.Consumed.GTE(quota.Allocated) {
			err := fmt.Errorf("quota exceeded; allocated %s, consumed %s", quota.Allocated, quota.Consumed)
			c.JSON(http.StatusBadRequest, types.NewResponseError(10, err))
			return
		}

		result, err := ctx.Service().AddPeer(req.key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, types.NewResponseError(11, err))
			return
		}
		ctx.Log().Info("Added a new peer", "key", key, "count", ctx.Service().PeersCount())

		ctx.Database().Model(
			&types.Session{},
		).Create(
			&types.Session{
				ID:           req.id,
				Subscription: subscription.Id,
				Key:          key,
				Address:      req.URI.AccAddress,
				Available:    quota.Allocated.Sub(quota.Consumed).Int64(),
			},
		)

		result = append(result, net.ParseIP(ctx.Location().IP).To4()...)
		result = append(result, ctx.Service().Info()...)
		c.JSON(http.StatusCreated, types.NewResponseResult(result))
	}
}
