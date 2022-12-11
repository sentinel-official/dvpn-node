package status

import (
	"time"
)

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
	QOS struct {
		MaxPeers int `json:"max_peers"`
	}
)

type (
	ResponseGetStatus struct {
		Address                string        `json:"address"`
		Bandwidth              *Bandwidth    `json:"bandwidth"`
		Handshake              *Handshake    `json:"handshake"`
		IntervalSetSessions    time.Duration `json:"interval_set_sessions"`
		IntervalUpdateSessions time.Duration `json:"interval_update_sessions"`
		IntervalUpdateStatus   time.Duration `json:"interval_update_status"`
		Location               *Location     `json:"location"`
		Moniker                string        `json:"moniker"`
		Operator               string        `json:"operator"`
		Peers                  int           `json:"peers"`
		Price                  string        `json:"price"`
		Provider               string        `json:"provider"`
		QOS                    *QOS          `json:"qos"`
		Type                   uint64        `json:"type"`
		Version                string        `json:"version"`
	}
)
