package types

import (
	"fmt"
)

type Request struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      string `json:"id"`
	Params  struct {
		Query string `json:"query"`
	} `json:"params"`
}

func NewRequest() Request {
	return Request{}
}

func NewDefaultRequest() Request {
	return Request{}.
		WithJSONRPC("2.0").
		WithID("0").
		WithMethod("subscribe")
}

func (r Request) WithJSONRPC(jsonRPC string) Request {
	r.JSONRPC = jsonRPC

	return r
}

func (r Request) WithID(id string) Request {
	r.ID = id

	return r
}

func (r Request) WithMethod(method string) Request {
	r.Method = method

	return r
}

func (r Request) WithQuery(query string) Request {
	r.Params.Query = query

	return r
}

func NewTxRequest(hash string) Request {
	query := fmt.Sprintf("tm.event = 'Tx' AND tx.hash = '%s'", hash)

	return NewDefaultRequest().
		WithQuery(query)
}
