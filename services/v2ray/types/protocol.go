package types

import "github.com/v2fly/v2ray-core/v4/app/proxyman/command"

type Protocol interface {
	Tag() string
	AddPeer(hsClient command.HandlerServiceClient, data []byte) error
	RemovePeer(hsClient command.HandlerServiceClient, data []byte) error
}
