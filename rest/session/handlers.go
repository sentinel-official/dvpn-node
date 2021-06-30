package session

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
	hubtypes "github.com/sentinel-official/hub/types"

	"github.com/sentinel-official/dvpn-node/context"
	"github.com/sentinel-official/dvpn-node/types"
	"github.com/sentinel-official/dvpn-node/utils"
)

func handlerAddSession(ctx *context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := NewRequestAddSession(r)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 1, err.Error())
			return
		}
		if err := body.Validate(); err != nil {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 1, err.Error())
			return
		}

		var (
			vars         = mux.Vars(r)
			key, _       = base64.StdEncoding.DecodeString(body.Key)
			signature, _ = base64.StdEncoding.DecodeString(body.Signature)
		)

		address, err := sdk.AccAddressFromBech32(vars["address"])
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 1, err.Error())
			return
		}

		id, err := strconv.ParseUint(vars["id"], 10, 64)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 1, err.Error())
			return
		}

		account, err := ctx.Client().QueryAccount(address)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 2, err.Error())
			return
		}
		if account == nil {
			err := fmt.Errorf("account %s does not exist", address)
			utils.WriteErrorToResponse(w, http.StatusNotFound, 2, err.Error())
			return
		}
		if account.GetPubKey() == nil {
			err := fmt.Errorf("public key for %s does not exist", address)
			utils.WriteErrorToResponse(w, http.StatusNotFound, 2, err.Error())
			return
		}
		if ok := account.GetPubKey().VerifySignature(sdk.Uint64ToBigEndian(id), signature); !ok {
			err := fmt.Errorf("failed to verify the signature %s", signature)
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 2, err.Error())
			return
		}

		session, err := ctx.Client().QuerySession(id)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 3, err.Error())
			return
		}
		if session == nil {
			err := fmt.Errorf("session %d does not exist", id)
			utils.WriteErrorToResponse(w, http.StatusNotFound, 3, err.Error())
			return
		}
		if !session.Status.Equal(hubtypes.StatusActive) {
			err := fmt.Errorf("invalid status %s for session %d", session.Status, session.Id)
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 3, err.Error())
			return
		}
		if session.Address != address.String() {
			err := fmt.Errorf("account address mismatch; expected %s, got %s", address, session.Address)
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 3, err.Error())
			return
		}

		subscription, err := ctx.Client().QuerySubscription(session.Subscription)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 4, err.Error())
			return
		}
		if subscription == nil {
			err := fmt.Errorf("subscription %d does not exist", session.Subscription)
			utils.WriteErrorToResponse(w, http.StatusNotFound, 4, err.Error())
			return
		}
		if !subscription.Status.Equal(hubtypes.Active) {
			err := fmt.Errorf("invalid status %s for subscription %d", subscription.Status, subscription.Id)
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 4, err.Error())
			return
		}

		if subscription.Plan == 0 {
			if subscription.Node != ctx.Address().String() {
				err := fmt.Errorf("node address mismatch; expected %s, got %s", ctx.Address(), subscription.Node)
				utils.WriteErrorToResponse(w, http.StatusBadRequest, 5, err.Error())
				return
			}
		} else {
			ok, err := ctx.Client().HasNodeForPlan(subscription.Plan, ctx.Address())
			if err != nil {
				utils.WriteErrorToResponse(w, http.StatusInternalServerError, 5, err.Error())
				return
			}
			if !ok {
				err := fmt.Errorf("node %s does not exist for plan %d", ctx.Address(), id)
				utils.WriteErrorToResponse(w, http.StatusBadRequest, 5, err.Error())
				return
			}
		}

		var item types.Session
		ctx.Database().Model(
			&types.Session{},
		).Where(
			&types.Session{
				ID: id,
			},
		).First(&item)

		if item.ID != 0 {
			err := fmt.Errorf("peer for session %d already exist", id)
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 6, err.Error())
			return
		}

		item = types.Session{}
		ctx.Database().Model(
			&types.Session{},
		).Where(
			&types.Session{
				Key: body.Key,
			},
		).First(&item)

		if item.ID != 0 {
			err := fmt.Errorf("key %s for service already exist", body.Key)
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 6, err.Error())
			return
		}

		var items []types.Session
		ctx.Database().Model(
			&types.Session{},
		).Where(
			&types.Session{
				Subscription: subscription.Id,
				Address:      address.String(),
			},
		).Find(&items)

		for i := 0; i < len(items); i++ {
			session, err := ctx.Client().QuerySession(items[i].ID)
			if err != nil {
				utils.WriteErrorToResponse(w, http.StatusInternalServerError, 7, err.Error())
				return
			}
			if session == nil {
				err := fmt.Errorf("session %d does not exist", items[i].ID)
				utils.WriteErrorToResponse(w, http.StatusNotFound, 7, err.Error())
				return
			}
			if session.Status.Equal(hubtypes.StatusActive) {
				err := fmt.Errorf("invalid status %s for session %d", session.Status, session.Id)
				utils.WriteErrorToResponse(w, http.StatusBadRequest, 7, err.Error())
				return
			}

			if err := ctx.RemovePeer(items[i].Key); err != nil {
				utils.WriteErrorToResponse(w, http.StatusInternalServerError, 8, err.Error())
				return
			}

			if session.Status.Equal(hubtypes.StatusInactive) {
				ctx.Database().Model(
					&types.Session{},
				).Where(
					&types.Session{
						ID: items[i].ID,
					},
				).Delete(
					&types.Session{},
				)
			}
		}

		quota, err := ctx.Client().QueryQuota(subscription.Id, address)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 9, err.Error())
			return
		}
		if quota == nil {
			err := fmt.Errorf("quota for address %s does not exist", address)
			utils.WriteErrorToResponse(w, http.StatusNotFound, 9, err.Error())
			return
		}

		items = []types.Session{}
		ctx.Database().Model(
			&types.Session{},
		).Where(
			&types.Session{
				Subscription: subscription.Id,
				Address:      address.String(),
			},
		).Find(&items)

		for i := 0; i < len(items); i++ {
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
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 10, err.Error())
			return
		}

		result, err := ctx.Service().AddPeer(key)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 11, err.Error())
			return
		}
		ctx.Log().Info("Added a new peer", "key", body.Key, "count", ctx.Service().PeersCount())

		ctx.Database().Model(
			&types.Session{},
		).Create(
			&types.Session{
				ID:           id,
				Subscription: subscription.Id,
				Key:          body.Key,
				Address:      address.String(),
				Available:    quota.Allocated.Sub(quota.Consumed).Int64(),
			},
		)

		result = append(result, net.ParseIP(ctx.Location().IP).To4()...)
		result = append(result, ctx.Service().Info()...)
		utils.WriteResultToResponse(w, http.StatusCreated, result)
	}
}
