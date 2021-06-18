package protocols

import (
	"context"
	"github.com/v2fly/v2ray-core/v4/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v4/common/protocol"
	"github.com/v2fly/v2ray-core/v4/common/serial"
	"github.com/v2fly/v2ray-core/v4/proxy/vmess"
)

type VMess struct {}

func (v VMess) Tag() string {
	return "vmess"
}

func (v VMess) AddPeer(hsClient command.HandlerServiceClient, data []byte) error {
	email := string(data)
	password := email[:36]

	_, err := hsClient.AlterInbound(context.Background(), &command.AlterInboundRequest{
		Tag: v.Tag(),
		Operation: serial.ToTypedMessage(
			&command.AddUserOperation{
				User: &protocol.User{
					Email: email,
					Account: serial.ToTypedMessage(&vmess.Account{
						Id: password,
					}),
				},
			}),
	})
	return err
}

func (v VMess) RemovePeer(hsClient command.HandlerServiceClient, data []byte) error {
	email := string(data)
	_, err := hsClient.AlterInbound(context.Background(), &command.AlterInboundRequest{
		Tag:       v.Tag(),
		Operation: serial.ToTypedMessage(&command.RemoveUserOperation{Email: email}),
	})
	return err
}
