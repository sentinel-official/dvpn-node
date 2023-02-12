package types

type Transport byte

func (t Transport) Byte() byte {
	return byte(t)
}

func (t Transport) IsValid() bool {
	return t.String() != ""
}

func (t Transport) String() string {
	switch t.Byte() {
	case 0x01:
		return "tcp"
	case 0x02:
		return "mkcp"
	case 0x03:
		return "websocket"
	case 0x04:
		return "http"
	case 0x05:
		return "domainsocket"
	case 0x06:
		return "quic"
	case 0x07:
		return "gun"
	case 0x08:
		return "grpc"
	default:
		return ""
	}
}

func NewTransportFromString(v string) Transport {
	switch v {
	case "tcp":
		return 0x01
	case "mkcp":
		return 0x02
	case "websocket":
		return 0x03
	case "httpt":
		return 0x04
	case "domainsocket":
		return 0x05
	case "quic":
		return 0x06
	case "gun":
		return 0x07
	case "grpc":
		return 0x08
	default:
		return 0x00
	}
}
