package metrics

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/version"
	"github.com/sentinel-official/dvpn-node/context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	nodeInfoGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dvpn_node_info",
			Help: "dVPN node info",
		},
		[]string{"address", "moniker", "operator", "version"},
	)

	locationInfoGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dvpn_location_info",
			Help: "dVPN location info",
		},
		[]string{"address", "moniker", "operator", "city", "country", "latitude", "longitude"},
	)

	handshakeEnabledGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dvpn_handshake_enabled",
			Help: "Is Handshake enabled for dVPN node?",
		},
		[]string{"address", "moniker", "operator"},
	)

	handshakePeersGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dvpn_handshake_peers",
			Help: "dVPN node Handshake peers",
		},
		[]string{"address", "moniker", "operator"},
	)

	bandwidthDownloadGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dvpn_bandwidth_download",
			Help: "dVPN node bandwidth download",
		},
		[]string{"address", "moniker", "operator"},
	)

	bandwidthUploadGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dvpn_bandwidth_upwnload",
			Help: "dVPN node bandwidth download",
		},
		[]string{"address", "moniker", "operator"},
	)

	intervalSetSessionsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dvpn_interval_set_sessions",
			Help: "dVPN node set session interval",
		},
		[]string{"address", "moniker", "operator"},
	)

	intervalUpdateSessionsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dvpn_interval_update_sessions",
			Help: "dVPN node update session interval",
		},
		[]string{"address", "moniker", "operator"},
	)

	intervalUpdateStatusGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dvpn_interval_update_status",
			Help: "dVPN node update status interval",
		},
		[]string{"address", "moniker", "operator"},
	)

	peersConnectedGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dvpn_peers_connected",
			Help: "dVPN peers connected",
		},
		[]string{"address", "moniker", "operator"},
	)

	peersMaxGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dvpn_peers_max",
			Help: "dVPN peers max",
		},
		[]string{"address", "moniker", "operator"},
	)

	priceGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dvpn_price",
			Help: "dVPN price per GB",
		},
		[]string{"address", "moniker", "operator", "denom"},
	)

	registry = prometheus.NewRegistry()
)

func HandlerGetMetrics(ctx *context.Context) http.HandlerFunc {
	registry.MustRegister(nodeInfoGauge)
	registry.MustRegister(locationInfoGauge)
	registry.MustRegister(handshakeEnabledGauge)
	registry.MustRegister(handshakePeersGauge)
	registry.MustRegister(bandwidthDownloadGauge)
	registry.MustRegister(bandwidthUploadGauge)
	registry.MustRegister(intervalSetSessionsGauge)
	registry.MustRegister(intervalUpdateSessionsGauge)
	registry.MustRegister(intervalUpdateStatusGauge)
	registry.MustRegister(peersConnectedGauge)

	return func(w http.ResponseWriter, r *http.Request) {
		nodeInfoGauge.With(prometheus.Labels{
			"address":  ctx.Address().String(),
			"moniker":  ctx.Moniker(),
			"operator": ctx.Operator().String(),
			"version":  version.Version,
		}).Set(1)

		locationInfoGauge.With(prometheus.Labels{
			"address":   ctx.Address().String(),
			"moniker":   ctx.Moniker(),
			"operator":  ctx.Operator().String(),
			"city":      ctx.Location().City,
			"country":   ctx.Location().Country,
			"latitude":  fmt.Sprintf("%f", ctx.Location().Latitude),
			"longitude": fmt.Sprintf("%f", ctx.Location().Longitude),
		}).Set(1)

		var handshakeEnabled float64 = 0
		if ctx.Config().Handshake.Enable {
			handshakeEnabled = 1
		}

		handshakeEnabledGauge.With(prometheus.Labels{
			"address":  ctx.Address().String(),
			"moniker":  ctx.Moniker(),
			"operator": ctx.Operator().String(),
		}).Set(handshakeEnabled)

		handshakePeersGauge.With(prometheus.Labels{
			"address":  ctx.Address().String(),
			"moniker":  ctx.Moniker(),
			"operator": ctx.Operator().String(),
		}).Set(float64(ctx.Config().Handshake.Peers))

		bandwidthDownloadGauge.With(prometheus.Labels{
			"address":  ctx.Address().String(),
			"moniker":  ctx.Moniker(),
			"operator": ctx.Operator().String(),
		}).Set(float64(ctx.Bandwidth().Download.Int64()))

		bandwidthUploadGauge.With(prometheus.Labels{
			"address":  ctx.Address().String(),
			"moniker":  ctx.Moniker(),
			"operator": ctx.Operator().String(),
		}).Set(float64(ctx.Bandwidth().Upload.Int64()))

		intervalSetSessionsGauge.With(prometheus.Labels{
			"address":  ctx.Address().String(),
			"moniker":  ctx.Moniker(),
			"operator": ctx.Operator().String(),
		}).Set(ctx.IntervalSetSessions().Seconds())

		intervalUpdateSessionsGauge.With(prometheus.Labels{
			"address":  ctx.Address().String(),
			"moniker":  ctx.Moniker(),
			"operator": ctx.Operator().String(),
		}).Set(ctx.IntervalUpdateSessions().Seconds())

		intervalUpdateStatusGauge.With(prometheus.Labels{
			"address":  ctx.Address().String(),
			"moniker":  ctx.Moniker(),
			"operator": ctx.Operator().String(),
		}).Set(ctx.IntervalUpdateStatus().Seconds())

		peersConnectedGauge.With(prometheus.Labels{
			"address":  ctx.Address().String(),
			"moniker":  ctx.Moniker(),
			"operator": ctx.Operator().String(),
		}).Set(float64(ctx.Service().PeersCount()))

		peersMaxGauge.With(prometheus.Labels{
			"address":  ctx.Address().String(),
			"moniker":  ctx.Moniker(),
			"operator": ctx.Operator().String(),
		}).Set(float64(ctx.Config().QOS.MaxPeers))

		for _, coin := range ctx.Price() {
			priceGauge.With(prometheus.Labels{
				"address":  ctx.Address().String(),
				"moniker":  ctx.Moniker(),
				"operator": ctx.Operator().String(),
				"denom":    coin.Denom,
			}).Set(float64(coin.Amount.Int64()))
		}

		promhttp.HandlerFor(
			registry,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		).ServeHTTP(w, r)
	}
}
