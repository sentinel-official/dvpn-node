package server

import (
	"encoding/json"
	"net/http"
)

func sendErrorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(message); err != nil {
		panic(err)
	}
}
