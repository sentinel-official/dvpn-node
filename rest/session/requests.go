package session

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
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
		return errors.New("key cannot be empty")
	}
	if _, err := base64.StdEncoding.DecodeString(r.Key); err != nil {
		return errors.Wrap(err, "invalid key")
	}
	if r.Signature == "" {
		return errors.New("signature cannot be empty")
	}
	if _, err := base64.StdEncoding.DecodeString(r.Signature); err != nil {
		return errors.Wrap(err, "invalid signature")
	}

	return nil
}
