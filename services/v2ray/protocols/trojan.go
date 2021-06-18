package protocols

import (
	"context"
	"github.com/v2fly/v2ray-core/v4/app/proxyman/command"
	"github.com/v2fly/v2ray-core/v4/common/protocol"
	"github.com/v2fly/v2ray-core/v4/common/serial"
	"github.com/v2fly/v2ray-core/v4/proxy/trojan"
)

type Trojan struct {}

func (t Trojan) Tag() string {
	return "trojan"
}

func (t Trojan) AddPeer(hsClient command.HandlerServiceClient, data []byte) error {
	email := string(data)
	password := email[:36]

	_, err := hsClient.AlterInbound(context.Background(), &command.AlterInboundRequest{
		Tag: t.Tag(),
		Operation: serial.ToTypedMessage(
			&command.AddUserOperation{
				User: &protocol.User{
					Email: email,
					Account: serial.ToTypedMessage(&trojan.Account{
						Password: password,
					}),
				},
			}),
	})
	return err
}

func (t Trojan) RemovePeer(hsClient command.HandlerServiceClient, data []byte) error {
	email := string(data)
	_, err := hsClient.AlterInbound(context.Background(), &command.AlterInboundRequest{
		Tag:       t.Tag(),
		Operation: serial.ToTypedMessage(&command.RemoveUserOperation{Email: email}),
	})
	return err
}
