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
