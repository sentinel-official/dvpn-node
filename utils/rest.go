package utils

import (
	"encoding/json"
	"net/http"

	"github.com/sentinel-official/dvpn-node/types"
)

func WriteErrorToResponse(w http.ResponseWriter, status, code int, message string) {
	_ = write(w, status, types.Response{
		Success: false,
		Error: struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{
			Code:    code,
			Message: message,
		},
	})
}

func WriteResultToResponse(w http.ResponseWriter, status int, result interface{}) {
	_ = write(w, status, types.Response{
		Success: true,
		Result:  result,
	})
}

func write(w http.ResponseWriter, status int, res types.Response) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(res)
}
