package utils

import (
	"encoding/json"
	"net/http"

	"github.com/sentinel-official/dvpn-node/types"
)

func FetchGeoIPLocation() (*types.GeoIPLocation, error) {
	resp, err := http.Get("http://ip-api.com/json")
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	var result struct {
		City    string  `json:"city"`
		Country string  `json:"country"`
		Lat     float64 `json:"lat"`
		Lon     float64 `json:"lon"`
		Query   string  `json:"query"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &types.GeoIPLocation{
		City:      result.City,
		Country:   result.Country,
		IP:        result.Query,
		Latitude:  result.Lat,
		Longitude: result.Lon,
	}, nil
}
