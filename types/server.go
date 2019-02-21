package types

import (
	"time"

	"github.com/gorilla/websocket"
)

var (
	Upgrader = &websocket.Upgrader{
		HandshakeTimeout: 45 * time.Second,
	}
)

type Client struct {
	Conn        *websocket.Conn
	OutMessages chan []byte
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		Conn:        conn,
		OutMessages: make(chan []byte),
	}
}
