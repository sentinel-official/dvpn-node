package types

import (
	"github.com/sentinel-official/dvpn-node/utils"
)

const (
	Type           = 2
	ConfigFileName = "v2ray.toml"
)

var (
	TransportProtocols = utils.SortStrings(
		[]string{
			"tcp",
			"mkcp",
			"websocket",
			"http",
			"domainsocket",
			"quic",
			"gun",
			"grpc",
		},
	)
)
