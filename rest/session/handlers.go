package session

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

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
			utils.WriteErrorToResponse(w, http.StatusNotFound, 2, "account does not exist")
			return
		}
		if account.GetPubKey() == nil {
			utils.WriteErrorToResponse(w, http.StatusNotFound, 2, "public key does not exist")
			return
		}
		if ok := account.GetPubKey().VerifySignature(sdk.Uint64ToBigEndian(id), signature); !ok {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 2, "failed to verify the signature")
			return
		}

		item := ctx.Sessions().GetByAddress(address)
		if item.Empty() {
			item = ctx.Sessions().GetByKey(body.Key)
		}

		if !item.Empty() {
			session, err := ctx.Client().QuerySession(item.ID)
			if err != nil {
				utils.WriteErrorToResponse(w, http.StatusInternalServerError, 3, err.Error())
				return
			}
			if session == nil {
				utils.WriteErrorToResponse(w, http.StatusNotFound, 3, "session does not exist")
				return
			}
			if session.Status.Equal(hubtypes.StatusActive) {
				utils.WriteErrorToResponse(w, http.StatusBadRequest, 3, fmt.Sprintf("invalid session status %s", session.Status))
				return
			}

			if err := ctx.RemovePeerAndSession(item.Key, item.Address); err != nil {
				utils.WriteErrorToResponse(w, http.StatusInternalServerError, 3, err.Error())
				return
			}

			if session.Status.Equal(hubtypes.StatusInactivePending) {
				go func() {
					_ = ctx.UpdateSessions(item)
				}()
			}
		}

		session, err := ctx.Client().QuerySession(id)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 4, err.Error())
			return
		}
		if session == nil {
			utils.WriteErrorToResponse(w, http.StatusNotFound, 4, "session does not exist")
			return
		}
		if !session.Status.Equal(hubtypes.StatusActive) {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 4, fmt.Sprintf("invalid session status %s", session.Status))
			return
		}
		if session.Address != address.String() {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 4, "account address mismatch")
			return
		}

		subscription, err := ctx.Client().QuerySubscription(session.Subscription)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 5, err.Error())
			return
		}
		if subscription == nil {
			utils.WriteErrorToResponse(w, http.StatusNotFound, 5, "subscription does not exist")
			return
		}
		if !subscription.Status.Equal(hubtypes.Active) {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 5, fmt.Sprintf("invalid subscription status %s", subscription.Status))
			return
		}

		if subscription.Plan == 0 {
			if subscription.Node != ctx.Address().String() {
				utils.WriteErrorToResponse(w, http.StatusBadRequest, 6, "node address mismatch")
				return
			}
		} else {
			ok, err := ctx.Client().HasNodeForPlan(id, ctx.Address())
			if err != nil {
				utils.WriteErrorToResponse(w, http.StatusInternalServerError, 6, err.Error())
				return
			}
			if !ok {
				utils.WriteErrorToResponse(w, http.StatusBadRequest, 6, "node does not exist for plan")
				return
			}
		}

		quota, err := ctx.Client().QueryQuota(subscription.Id, address)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 7, err.Error())
			return
		}
		if quota == nil {
			utils.WriteErrorToResponse(w, http.StatusNotFound, 7, "quota does not exist")
			return
		}
		if quota.Consumed.GTE(quota.Allocated) {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 7, "quota exceeded")
			return
		}

		result, err := ctx.Service().AddPeer(key)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 8, err.Error())
			return
		}
		ctx.Log().Info("Added a new peer", "key", body.Key, "count", ctx.Service().PeersCount())

		ctx.Sessions().Set(
			types.Session{
				ID:          id,
				Key:         body.Key,
				Address:     address,
				Available:   quota.Allocated.Sub(quota.Consumed),
				ConnectedAt: time.Now(),
			},
		)
		ctx.Log().Info("Added a new session", "id", id, "address", address, "count", ctx.Sessions().Len())

		result = append(result, net.ParseIP(ctx.Location().IP).To4()...)
		result = append(result, ctx.Service().Info()...)
		utils.WriteResultToResponse(w, http.StatusCreated, result)
	}
}
