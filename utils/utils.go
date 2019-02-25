package utils

import (
	"encoding/json"
	"net/http"
)

func PublicIP() (string, error) {
	r, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		return "", err
	}

	var res struct {
		IP string
	}

	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		return "", err
	}

	return res.IP, nil
}
