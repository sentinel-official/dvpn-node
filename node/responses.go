package node

import (
	"encoding/json"
	"net/http"

	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	Success bool        `json:"success"`
	Error   Error       `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

func writeErrorToResponse(w http.ResponseWriter, statusCode int, _error Error) {
	res := Response{
		Success: false,
		Error:   _error,
	}

	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		panic(err)
	}
}

func writeResultToResponse(w http.ResponseWriter, statusCode int, result interface{}) {
	res := Response{
		Success: true,
		Result:  result,
	}

	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		panic(err)
	}
}

type resultSessionKey struct {
	Key             string             `json:"key"`
	Bandwidth       sdkTypes.Bandwidth `json:"bandwidth"`
	ClientSignature string             `json:"client_signature"`
}
