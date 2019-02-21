package types

import (
	"encoding/json"
)

type Msg struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (m Msg) Bytes() []byte {
	data, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	return data
}
