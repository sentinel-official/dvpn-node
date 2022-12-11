package rest

import (
	"github.com/gorilla/mux"
	"github.com/sentinel-official/dvpn-node/context"
	"github.com/sentinel-official/dvpn-node/rest/session"
	"github.com/sentinel-official/dvpn-node/rest/status"
)

func RegisterRoutes(ctx *context.Context, r *mux.Router) {
	session.RegisterRoutes(ctx, r)
	status.RegisterRoutes(ctx, r)
}
