package metrics

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/sentinel-official/dvpn-node/context"
)

func RegisterRoutes(ctx *context.Context, r *mux.Router) {
	r.Name("GetMetrics").
		Methods(http.MethodGet).Path("/metrics").
		HandlerFunc(HandlerGetMetrics(ctx))
}
