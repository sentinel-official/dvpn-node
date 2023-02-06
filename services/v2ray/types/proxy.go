package types

import (
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/common/uuid"
	"github.com/v2fly/v2ray-core/v5/proxy/vmess"
	"google.golang.org/protobuf/types/known/anypb"
)

type Proxy byte

func (p Proxy) Byte() byte {
	return byte(p)
}

func (p Proxy) IsValid() bool {
	return p.String() != ""
}

func (p Proxy) Tag() string {
	return p.String()
}

func (p Proxy) String() string {
	switch p.Byte() {
	case 0x01:
		return "vmess"
	default:
		return ""
	}
}

func (p Proxy) Account(uid uuid.UUID) *anypb.Any {
	switch p.Byte() {
	case 0x01:
		return serial.ToTypedMessage(
			&vmess.Account{
				Id:      uid.String(),
				AlterId: 0,
				SecuritySettings: &protocol.SecurityConfig{
					Type: protocol.SecurityType_AUTO,
				},
				TestsEnabled: "",
			},
		)
	default:
		return nil
	}
}
