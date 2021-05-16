package session

import (
	"encoding/base64"
	"encoding/hex"
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
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 2, err.Error())
			return
		}

		var (
			vars = mux.Vars(r)
		)

		address, err := hex.DecodeString(vars["address"])
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 2, err.Error())
			return
		}

		id, err := strconv.ParseUint(vars["id"], 10, 64)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 2, err.Error())
			return
		}

		key, err := base64.StdEncoding.DecodeString(body.Key)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 2, err.Error())
			return
		}

		signature, err := base64.StdEncoding.DecodeString(body.Signature)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 2, err.Error())
			return
		}

		account, err := ctx.Client().QueryAccount(address)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 3, err.Error())
			return
		}
		if account == nil || account.GetPubKey() == nil {
			utils.WriteErrorToResponse(w, http.StatusNotFound, 3, "account does not exist")
			return
		}
		if ok := account.GetPubKey().VerifySignature(sdk.Uint64ToBigEndian(id), signature); !ok {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 3, "failed to verify signature")
			return
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
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 4, "invalid session status")
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
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 5, "invalid subscription status")
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

		quota, err := ctx.Client().QueryQuota(id, address)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 7, err.Error())
			return
		}
		if quota == nil {
			utils.WriteErrorToResponse(w, http.StatusNotFound, 7, "quota does not exist")
			return
		}
		if quota.Consumed.GTE(quota.Allocated) {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 7, "invalid quota")
			return
		}

		if ctx.Sessions().Get(body.Key).Key != "" {
			utils.WriteErrorToResponse(w, http.StatusConflict, 8, "duplicate key")
			return
		}

		ctx.Sessions().Put(
			types.Session{
				ID:          id,
				Key:         body.Key,
				Address:     address,
				Available:   quota.Allocated.Sub(quota.Consumed),
				ConnectedAt: time.Now(),
			},
		)

		result, err := ctx.Service().AddPeer(key)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 9, err.Error())
			return
		}

		result = append(result, net.ParseIP(ctx.Location().IP).To4()...)
		result = append(result, ctx.Service().Info()...)
		utils.WriteResultToResponse(w, http.StatusCreated, result)
	}
}
