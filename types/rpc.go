package types

type RPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      string `json:"id"`
	Params  struct {
		Query string `json:"query"`
	} `json:"params"`
}

func NewRPCRequest() RPCRequest {
	return RPCRequest{}
}

func (r RPCRequest) WithJSONRPC(jsonRPC string) RPCRequest {
	r.JSONRPC = jsonRPC

	return r
}

func (r RPCRequest) WithID(id string) RPCRequest {
	r.ID = id

	return r
}

func (r RPCRequest) WithMethod(method string) RPCRequest {
	r.Method = method

	return r
}

func (r RPCRequest) WithQuery(query string) RPCRequest {
	r.Params.Query = query

	return r
}
