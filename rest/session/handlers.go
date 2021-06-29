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
			var err = fmt.Errorf("account %s does not exist", address)
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

		var item types.Session
		ctx.Database().Where(
			&types.Session{
				Address: address.String(),
			},
		).First(&item)

		if item.Key == body.Key {
			err := fmt.Errorf("key %s for service already exist", body.Key)
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 3, err.Error())
			return
		}

		if item.ID != 0 {
			session, err := ctx.Client().QuerySession(item.ID)
			if err != nil {
				utils.WriteErrorToResponse(w, http.StatusInternalServerError, 4, err.Error())
				return
			}
			if session == nil {
				err := fmt.Errorf("session %d does not exist", item.ID)
				utils.WriteErrorToResponse(w, http.StatusNotFound, 4, err.Error())
				return
			}
			if session.Status.Equal(hubtypes.StatusActive) {
				err := fmt.Errorf("invalid session status %s", session.Status)
				utils.WriteErrorToResponse(w, http.StatusBadRequest, 4, err.Error())
				return
			}

			if err := ctx.RemovePeer(item.Key); err != nil {
				utils.WriteErrorToResponse(w, http.StatusInternalServerError, 4, err.Error())
				return
			}

			if session.Status.Equal(hubtypes.StatusInactive) {
				ctx.Database().Where(
					&types.Session{
						Address: item.Address,
					},
				).Delete(
					&types.Session{},
				)
			}
		}

		session, err := ctx.Client().QuerySession(id)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 5, err.Error())
			return
		}
		if session == nil {
			err := fmt.Errorf("session %d does not exist", id)
			utils.WriteErrorToResponse(w, http.StatusNotFound, 5, err.Error())
			return
		}
		if !session.Status.Equal(hubtypes.StatusActive) {
			err := fmt.Errorf("invalid session status %s", session.Status)
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 5, err.Error())
			return
		}
		if session.Address != address.String() {
			err := fmt.Errorf("account address mismatch; expected %s, got %s", address, session.Address)
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 5, err.Error())
			return
		}

		subscription, err := ctx.Client().QuerySubscription(session.Subscription)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 6, err.Error())
			return
		}
		if subscription == nil {
			err := fmt.Errorf("subscription %d does not exist", session.Subscription)
			utils.WriteErrorToResponse(w, http.StatusNotFound, 6, err.Error())
			return
		}
		if !subscription.Status.Equal(hubtypes.Active) {
			err := fmt.Errorf("invalid subscription status %s", subscription.Status)
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 6, err.Error())
			return
		}

		if subscription.Plan == 0 {
			if subscription.Node != ctx.Address().String() {
				err := fmt.Errorf("node address mismatch; got %s", subscription.Node)
				utils.WriteErrorToResponse(w, http.StatusBadRequest, 7, err.Error())
				return
			}
		} else {
			ok, err := ctx.Client().HasNodeForPlan(subscription.Plan, ctx.Address())
			if err != nil {
				utils.WriteErrorToResponse(w, http.StatusInternalServerError, 7, err.Error())
				return
			}
			if !ok {
				err := fmt.Errorf("node %s does not exist for plan %d", ctx.Address(), id)
				utils.WriteErrorToResponse(w, http.StatusBadRequest, 7, err.Error())
				return
			}
		}

		quota, err := ctx.Client().QueryQuota(subscription.Id, address)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 8, err.Error())
			return
		}
		if quota == nil {
			err := fmt.Errorf("quota for address %s does not exist", address)
			utils.WriteErrorToResponse(w, http.StatusNotFound, 8, err.Error())
			return
		}

		item = types.Session{}
		ctx.Database().Where(
			&types.Session{
				Address: address.String(),
			},
		).First(&item)

		if item.ID != 0 {
			quota.Consumed = quota.Consumed.Add(
				hubtypes.NewBandwidthFromInt64(
					item.Download, item.Upload,
				).CeilTo(
					hubtypes.Gigabyte.Quo(subscription.Price.Amount),
				).Sum(),
			)
		}

		if quota.Consumed.GTE(quota.Allocated) {
			err := fmt.Errorf("quota exceeded; consumed %d", quota.Consumed.Int64())
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 9, err.Error())
			return
		}

		result, err := ctx.Service().AddPeer(key)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 8, err.Error())
			return
		}
		ctx.Log().Info("Added a new peer", "key", body.Key, "count", ctx.Service().PeersCount())

		ctx.Database().Create(
			&types.Session{
				ID:        id,
				Key:       body.Key,
				Address:   address.String(),
				Available: quota.Allocated.Sub(quota.Consumed).Int64(),
			},
		)

		result = append(result, net.ParseIP(ctx.Location().IP).To4()...)
		result = append(result, ctx.Service().Info()...)
		utils.WriteResultToResponse(w, http.StatusCreated, result)
	}
}
