package tx

import (
	"fmt"

	"github.com/ironman0x7b2/vpn-node/types"
)

func SubscribeRPCRequest() types.RPCRequest {
	req := types.NewRPCRequest().
		WithJSONRPC("2.0").
		WithID("0").
		WithMethod("subscribe")

	return req
}

func SubscribeRPCRequestWithQuery(query string) types.RPCRequest {
	req := SubscribeRPCRequest().
		WithQuery(query)

	return req
}

func NewTxSubscribeRPCRequest(txHash string) types.RPCRequest {
	query := fmt.Sprintf("tm.event = 'Tx' AND tx.hash = '%s'", txHash)
	req := SubscribeRPCRequestWithQuery(query)

	return req
}
