package rest

import (
	"github.com/gorilla/mux"

	"github.com/sentinel-official/dvpn-node/context"
)

func RegisterRoutes(ctx *context.Context, router *mux.Router) {
	router.
		Name("GetStatus").
		Methods("GET").
		Path("/status").
		HandlerFunc(getStatus(ctx))

	router.
		Name("AddSession").
		Methods("POST").
		Path("/sessions").
		HandlerFunc(addSession(ctx))
}
