package status

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sentinel-official/dvpn-node/context"
)

func RegisterRoutes(ctx *context.Context, r *mux.Router) {
	r.Name("GetStatus").
		Methods(http.MethodGet).Path("/status").
		HandlerFunc(HandlerGetStatus(ctx))
}
