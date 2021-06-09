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
)

type (
	ResponseGetStatus struct {
		Address          string        `json:"address"`
		Bandwidth        *Bandwidth    `json:"bandwidth"`
		Handshake        *Handshake    `json:"handshake"`
		IntervalSessions time.Duration `json:"interval_sessions"`
		IntervalStatus   time.Duration `json:"interval_status"`
		Location         *Location     `json:"location"`
		Moniker          string        `json:"moniker"`
		Operator         string        `json:"operator"`
		Peers            int           `json:"peers"`
		Price            string        `json:"price"`
		Provider         string        `json:"provider"`
		Type             uint64        `json:"type"`
		Version          string        `json:"version"`
	}
)
