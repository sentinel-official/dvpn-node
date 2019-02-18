package server

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	vpnTypes "github.com/ironman0x7b2/sentinel-sdk/x/vpn/types"
)

var (
	pongWait   = 60 * time.Second
	pingPeriod = pongWait * 9 / 10
)

func NewUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		HandshakeTimeout: 45 * time.Second,
	}
}

type MsgBandwidthSign struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	Upload    int64  `json:"upload"`
	Download  int64  `json:"download"`
	Sign      string `json:"sign"`
}

func NewMsgBandwidthSign(id, sessionID string, upload, download int64, sign string) MsgBandwidthSign {
	return MsgBandwidthSign{
		ID:        id,
		SessionID: sessionID,
		Upload:    upload,
		Download:  download,
		Sign:      sign,
	}
}

func (b MsgBandwidthSign) Validate() error {
	if len(b.ID) == 0 {
		return fmt.Errorf("id is empty")
	}
	if len(b.SessionID) == 0 {
		return fmt.Errorf("session_id is empty")
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

type MsgError struct {
	Code    int8   `json:"code"`
	Message string `json:"message"`
}

func NewMsgError(code int8, message string) MsgError {
	return MsgError{
		Code:    code,
		Message: message,
	}
}

type Connection struct {
	Conn        *websocket.Conn
	Session     *vpnTypes.SessionDetails
	OutMessages chan interface{}
}

func NewConnection(conn *websocket.Conn, session *vpnTypes.SessionDetails) *Connection {
	return &Connection{
		Conn:        conn,
		Session:     session,
		OutMessages: make(chan interface{}),
	}
}
