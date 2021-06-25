package utils

import (
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/transport/http/jsonrpc"

	"github.com/sentinel-official/dvpn-node/types"
)

func WriteErrorToResponse(w http.ResponseWriter, status, code int, message string) {
	_ = write(w, status, types.Response{
		Success: false,
		Error:   types.NewError("", code, message),
	})
}

func WriteResultToResponse(w http.ResponseWriter, status int, result interface{}) {
	_ = write(w, status, types.Response{
		Success: true,
		Result:  result,
	})
}

func write(w http.ResponseWriter, status int, res types.Response) error {
	w.Header().Set("Content-Type", jsonrpc.ContentType)
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(res)
}
