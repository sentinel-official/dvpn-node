package server

import (
	"time"

	"github.com/gorilla/websocket"
)

type webSocketReqBody struct {
	TxHash    string `json:"tx_hash"`
	Signature string `json:"signature"`
}

func NewUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		HandshakeTimeout: 45 * time.Second,
	}
}
