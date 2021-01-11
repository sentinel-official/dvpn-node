package session

import (
	"encoding/base64"
	"encoding/hex"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	hub "github.com/sentinel-official/hub/types"

	"github.com/sentinel-official/dvpn-node/context"
	"github.com/sentinel-official/dvpn-node/types"
	"github.com/sentinel-official/dvpn-node/utils"
)

func HandlerAddSession(ctx *context.Context) http.HandlerFunc {
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

		vars := mux.Vars(r)

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

		identity, err := base64.StdEncoding.DecodeString(body.Key)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 2, err.Error())
			return
		}

		subscription, err := ctx.Client().QuerySubscription(id)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 3, err.Error())
			return
		}
		if subscription == nil {
			utils.WriteErrorToResponse(w, http.StatusNotFound, 3, "")
			return
		}
		if !subscription.Status.Equal(hub.StatusActive) {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 3, "")
			return
		}

		if subscription.Plan == 0 {
			if !subscription.Node.Equals(ctx.Address()) {
				utils.WriteErrorToResponse(w, http.StatusBadRequest, 4, "")
				return
			}
		} else {
			ok, err := ctx.Client().HasNodeForPlan(id, ctx.Address())
			if err != nil {
				utils.WriteErrorToResponse(w, http.StatusInternalServerError, 4, err.Error())
				return
			}
			if !ok {
				utils.WriteErrorToResponse(w, http.StatusBadRequest, 4, "")
				return
			}
		}

		quota, err := ctx.Client().QueryQuota(id, address)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 5, err.Error())
			return
		}
		if quota == nil {
			utils.WriteErrorToResponse(w, http.StatusNotFound, 5, "")
			return
		}
		if quota.Consumed.GTE(quota.Allocated) {
			utils.WriteErrorToResponse(w, http.StatusBadRequest, 5, "")
			return
		}

		if ctx.Sessions().Get(body.Key).Identity != "" {
			utils.WriteErrorToResponse(w, http.StatusConflict, 6, "")
			return
		}

		ctx.Sessions().Set(types.Session{
			Address:      address,
			ConnectedAt:  time.Now(),
			Identity:     body.Key,
			Subscription: id,
		})

		result, err := ctx.Service().AddPeer(identity)
		if err != nil {
			utils.WriteErrorToResponse(w, http.StatusInternalServerError, 7, err.Error())
			return
		}

		result = append(result, net.ParseIP(ctx.Location().IP)...)
		result = append(result, ctx.Service().Info()...)
		utils.WriteResultToResponse(w, http.StatusCreated, result)
	}
}
