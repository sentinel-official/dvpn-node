package node

type requestAddSubscription struct {
	TxHash string `json:"tx_hash"`
}

type requestSubscriptionWebsocket struct {
	Signature string `json:"signature"`
}
