package status

import (
	"net/http"

	"github.com/sentinel-official/dvpn-node/context"
	"github.com/sentinel-official/dvpn-node/utils"
)

func HandlerGetStatus(ctx *context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.WriteResultToResponse(w, http.StatusOK, ResponseGetStatus{
			Address: ctx.Address().String(),
			Bandwidth: Bandwidth{
				Upload:   ctx.Bandwidth().Upload.Int64(),
				Download: ctx.Bandwidth().Download.Int64(),
			},
			Handshake: Handshake{
				Enable: ctx.Config().Handshake.Enable,
				Peers:  ctx.Config().Handshake.Peers,
			},
			IntervalSessions: ctx.IntervalSessions(),
			IntervalStatus:   ctx.IntervalStatus(),
			Location: Location{
				City:      ctx.Location().City,
				Country:   ctx.Location().Country,
				Latitude:  ctx.Location().Latitude,
				Longitude: ctx.Location().Longitude,
			},
			Moniker:  ctx.Moniker(),
			Operator: ctx.Operator().String(),
			Peers:    ctx.Service().PeersCount(),
			Price:    ctx.Price().String(),
			Provider: ctx.Provider().String(),
			Type:     ctx.Type(),
			Version:  ctx.Version(),
		})
	}
}
