package utils

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/pkg/errors"
)

func PublicIP() (net.IP, error) {
	resp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	var res struct {
		IP string `json:"ip"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	ip := net.ParseIP(res.IP)
	if ip == nil {
		return nil, errors.Errorf("IP address is empty")
	}

	return ip, nil
}
