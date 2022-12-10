package wireguard

import (
	"strings"
)

//nolint:lll
var (
	configTemplate = strings.TrimSpace(`
[Interface]
Address = 10.8.0.1/24,fd86:ea04:1115::1/120
ListenPort = {{ .ListenPort }}
PrivateKey = {{ .PrivateKey }}
PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE; ip6tables -A FORWARD -i %i -j ACCEPT; ip6tables -t nat -A POSTROUTING -o eth0 -j MASQUERADE;
PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE; ip6tables -D FORWARD -i %i -j ACCEPT; ip6tables -t nat -D POSTROUTING -o eth0 -j MASQUERADE;
    `)
)
