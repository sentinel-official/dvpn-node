package utils

import (
	"encoding/json"
	"net/http"

	"github.com/ironman0x7b2/vpn-node/types"
)

func WriteErrorToResponse(w http.ResponseWriter, statusCode int, err interface{}) {
	res := types.Response{
		Success: false,
		Error:   err,
	}

	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		panic(err)
	}
}

func WriteResultToResponse(w http.ResponseWriter, statusCode int, result interface{}) {
	res := types.Response{
		Success: true,
		Result:  result,
	}

	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		panic(err)
	}
}
