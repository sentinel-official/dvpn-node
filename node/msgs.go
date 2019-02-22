package node

import (
	"encoding/json"

	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
	"github.com/pkg/errors"

	"github.com/ironman0x7b2/vpn-node/types"
)

type MsgBandwidthSign struct {
	SessionID     string             `json:"session_id"`
	Bandwidth     sdkTypes.Bandwidth `json:"bandwidth"`
	NodeOwnerSign string             `json:"node_owner_sign"`
	ClientSign    string             `json:"client_sign"`
}

func NewMsgBandwidthSign(sessionID string, bandwidth sdkTypes.Bandwidth, nodeOwnerSign, clientSign string) *types.Msg {
	msg := MsgBandwidthSign{
		SessionID:     sessionID,
		Bandwidth:     bandwidth,
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
	if !msg.Bandwidth.IsPositive() {
		return errors.New("bandwidth is not positive")
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
