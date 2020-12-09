package rest

import (
	"encoding/json"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	hub "github.com/sentinel-official/hub/types"

	"github.com/sentinel-official/dvpn-node/context"
	"github.com/sentinel-official/dvpn-node/types"
	"github.com/sentinel-official/dvpn-node/utils"
)

func addSession(ctx *context.Context) http.HandlerFunc {
	type (
		Request struct {
			ID      uint64 `json:"id"`
			Address string `json:"address"`
			Key     string `json:"key"`
		}
	)

	return func(w http.ResponseWriter, r *http.Request) {
		var body Request
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 1, err.Error())
			return
		}

		address, err := sdk.AccAddressFromHex(body.Address)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 2, err.Error())
			return
		}

		subscription, err := ctx.Client().QuerySubscription(body.ID)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 3, err.Error())
			return
		}
		if !subscription.Status.Equal(hub.StatusActive) {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 4, "")
			return
		}

		if subscription.Plan == 0 {
			if !subscription.Node.Equals(ctx.Address()) {
				utils.WriteErrorToResponse(w, http.StatusBadRequest, 5, "")
				return
			}
		} else {
			ok, err := ctx.Client().HasNodeForPlan(body.ID, ctx.Address())
			if err != nil {
				utils.WriteErrorToResponse(w, http.StatusInternalServerError, 5, err.Error())
				return
			}
			if !ok {
				utils.WriteErrorToResponse(w, http.StatusBadRequest, 5, "")
				return
			}
		}

		quota, err := ctx.Client().QueryQuota(body.ID, address)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 6, err.Error())
			return
		}
		if quota.Consumed.GTE(quota.Allocated) {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 7, "")
			return
		}

		if ctx.Sessions().Get(body.Key).Identity != "" {
			utils.WriteErrorToResponse(w, http.StatusConflict, 8, "")
			return
		}

		ctx.Sessions().Set(types.Session{
			Address:      address,
			ConnectedAt:  time.Now(),
			Identity:     body.Key,
			Subscription: body.ID,
		})

		result, err := ctx.Service().AddPeer([]byte(body.Key))
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 9, err.Error())
			return
		}

		utils.WriteResultToResponse(w, 201, result)
	}
}
