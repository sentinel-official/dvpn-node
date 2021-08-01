// +build linux
// +build !openwrt

package types

import (
	"strings"
)

var (
	ConfigTemplate = strings.TrimSpace(`
[Interface]
Address = {{ .IPv4CIDR }},{{ .IPv6CIDR }}
ListenPort = {{ .ListenPort }}
PrivateKey = {{ .PrivateKey }}
    `)
)
