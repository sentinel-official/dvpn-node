package session

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/sentinel-official/dvpn-node/context"
)

func RegisterRoutes(ctx *context.Context, r *mux.Router) {
	r.Name("AddSession").
		Methods(http.MethodPost).Path("/accounts/{address}/subscriptions/{id}/sessions").
		HandlerFunc(HandlerAddSession(ctx))
}
