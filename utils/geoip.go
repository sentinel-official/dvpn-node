package utils

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/sentinel-official/dvpn-node/types"
)

func FetchGeoIPLocation() (*types.GeoIPLocation, error) {
	var (
		client = &http.Client{Timeout: 15 * time.Second}
		path   = "https://ipv4.geojs.io/v1/ip/geo.json"
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
		City      string `json:"city"`
		Country   string `json:"country"`
		IP        string `json:"ip"`
		Latitude  string `json:"latitude"`
		Longitude string `json:"longitude"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	latitude, err := strconv.ParseFloat(body.Latitude, 64)
	if err != nil {
		return nil, err
	}

	longitude, err := strconv.ParseFloat(body.Longitude, 64)
	if err != nil {
		return nil, err
	}

	return &types.GeoIPLocation{
		City:      body.City,
		Country:   body.Country,
		IP:        body.IP,
		Latitude:  latitude,
		Longitude: longitude,
	}, nil
}
