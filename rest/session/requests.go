package session

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type RequestAddSession struct {
	Key       string `json:"key"`
	Signature string `json:"signature"`
}

func NewRequestAddSession(r *http.Request) (*RequestAddSession, error) {
	var body RequestAddSession
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	return &body, nil
}

func (r *RequestAddSession) Validate() error {
	if r.Key == "" {
		return fmt.Errorf(`invalid field key; expected non-empty value`)
	}

	return nil
}
