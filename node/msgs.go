package node

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/ironman0x7b2/vpn-node/types"
)

type MsgBandwidthSign struct {
	SessionID     string `json:"session_id"`
	Upload        int64  `json:"upload"`
	Download      int64  `json:"download"`
	NodeOwnerSign string `json:"node_owner_sign"`
	ClientSign    string `json:"client_sign"`
}

func NewMsgBandwidthSign(sessionID string, upload, download int64, nodeOwnerSign, clientSign string) *types.Msg {
	msg := MsgBandwidthSign{
		SessionID:     sessionID,
		Upload:        upload,
		Download:      download,
		NodeOwnerSign: nodeOwnerSign,
		ClientSign:    clientSign,
	}
	data, _ := json.Marshal(msg)

	return &types.Msg{
		Type: msg.Type(),
		Data: data,
	}
}

func (msg MsgBandwidthSign) Type() string {
	return "msg_bandwidth_sign"
}

func (msg MsgBandwidthSign) Validate() error {
	if len(msg.SessionID) == 0 {
		return errors.New("session_id is empty")
	}
	if msg.Upload <= 0 {
		return errors.New("upload is less than or equal to zero")
	}
	if msg.Download <= 0 {
		return errors.New("download is less than or equal to zero")
	}
	if len(msg.NodeOwnerSign) == 0 {
		return errors.New("node_owner_sign is empty")
	}
	if len(msg.ClientSign) == 0 {
		return errors.New("client_sign is empty")
	}

	return nil
}

type MsgError struct {
	Code    int8   `json:"code"`
	Message string `json:"message"`
}

func NewMsgError(code int8, message string) *types.Msg {
	msg := MsgError{
		Code:    code,
		Message: message,
	}
	data, _ := json.Marshal(msg)

	return &types.Msg{
		Type: msg.Type(),
		Data: data,
	}
}

func (msg MsgError) Validate() error {
	return nil
}

func (msg MsgError) Type() string {
	return "msg_error"
}
