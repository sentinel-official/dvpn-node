package tx

import (
	"fmt"

	"github.com/ironman0x7b2/vpn-node/types"
)

func SubscriberRPCRequest() types.RPCRequest {
	req := types.NewRPCRequest().
		WithJSONRPC("2.0").
		WithID("0").
		WithMethod("subscribe")

	return req
}

func SubscriberRPCRequestWithQuery(query string) types.RPCRequest {
	req := SubscriberRPCRequest().
		WithQuery(query)

	return req
}

func NewTxSubscriberRPCRequest(txHash string) types.RPCRequest {
	query := fmt.Sprintf("tm.event = 'Tx' AND tx.hash = '%s'", txHash)
	req := SubscriberRPCRequestWithQuery(query)

	return req
}
