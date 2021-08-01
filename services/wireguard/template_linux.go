// +build linux
// +build !openwrt

package wireguard

import (
	"strings"
)

var (
	configTemplate = strings.TrimSpace(`
[Interface]
Address = {{ .IPv4CIDR }},{{ .IPv6CIDR }}
ListenPort = {{ .ListenPort }}
PrivateKey = {{ .PrivateKey }}
    `)
)
