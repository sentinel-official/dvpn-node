package server

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	vpnTypes "github.com/ironman0x7b2/sentinel-sdk/x/vpn/types"
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

type bandwidthSignMsg struct {
	ID       string `json:"id"`
	Index    int64  `json:"index"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
	Sign     string `json:"sign"`
}

func (b bandwidthSignMsg) Validate() error {
	if len(b.ID) == 0 {
		return fmt.Errorf("id is empty")
	}
	if b.Index < 0 {
		return fmt.Errorf("index is less than zero")
	}
	if b.Upload <= 0 {
		return fmt.Errorf("upload is less than or equal to zero")
	}
	if b.Download <= 0 {
		return fmt.Errorf("download is less than or equal to zero")
	}
	if len(b.Sign) == 0 {
		return fmt.Errorf("sign is empty")
	}

	return nil
}

type Connection struct {
	Conn        *websocket.Conn
	Session     *vpnTypes.SessionDetails
	OutMessages chan struct{}
}

func NewConnection(conn *websocket.Conn, session *vpnTypes.SessionDetails) *Connection {
	return &Connection{
		Conn:        conn,
		Session:     session,
		OutMessages: make(chan struct{}),
	}
}
