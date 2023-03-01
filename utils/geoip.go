package utils

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sentinel-official/dvpn-node/types"
)

func FetchGeoIPLocation() (*types.GeoIPLocation, error) {
	var (
		client = &http.Client{Timeout: 15 * time.Second}
		path   = "http://ip-api.com/json"
	)

	resp, err := client.Get(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	var body struct {
		City      string  `json:"city"`
		Country   string  `json:"country"`
		IP        string  `json:"query"`
		Latitude  float64 `json:"lat"`
		Longitude float64 `json:"lon"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return &types.GeoIPLocation{
		City:      body.City,
		Country:   body.Country,
		IP:        body.IP,
		Latitude:  body.Latitude,
		Longitude: body.Longitude,
	}, nil
}
