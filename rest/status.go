package rest

import (
	"net/http"
	"time"

	"github.com/sentinel-official/dvpn-node/context"
	"github.com/sentinel-official/dvpn-node/utils"
)

func getStatus(ctx *context.Context) http.HandlerFunc {
	type (
		Bandwidth struct {
			Download int64 `json:"download"`
			Upload   int64 `json:"upload"`
		}
		Handshake struct {
			Enable bool   `json:"enable"`
			Peers  uint64 `json:"peers"`
		}
		Location struct {
			City      string  `json:"city"`
			Country   string  `json:"country"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		}
		Response struct {
			Address          string        `json:"address"`
			Bandwidth        Bandwidth     `json:"bandwidth"`
			Handshake        Handshake     `json:"handshake"`
			IntervalSessions time.Duration `json:"interval_sessions"`
			IntervalStatus   time.Duration `json:"interval_status"`
			Location         Location      `json:"location"`
			Moniker          string        `json:"moniker"`
			Operator         string        `json:"operator"`
			Peers            int64         `json:"peers"`
			Price            string        `json:"price"`
			Provider         string        `json:"provider"`
			Type             string        `json:"type"`
			Version          string        `json:"version"`
		}
	)

	return func(w http.ResponseWriter, r *http.Request) {
		utils.WriteResultToResponse(w, http.StatusOK, Response{
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
			Type:     ctx.Type().String(),
			Version:  ctx.Version(),
		})
	}
}
