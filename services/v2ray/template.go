package v2ray

import (
	"strings"
)

// nolint:lll
var (
	configTemplate = strings.TrimSpace(`
{
	"log": {
		"logLevel": "debug"
	},
	"stats": {},
	"policy": {
		"levels": {
			"0": {
				"statsUserUplink": true,
				"statsUserDownlink": true
			}
		},
		"system": {
			"statsInboundUplink": true,
			"statsInboundDownlink": true,
			"statsOutboundUplink": true,
			"statsOutboundDownlink": true
		}
	},
	"api": {
		"tag": "api",
		"services": ["HandlerService", "StatsService"]
	},
	"inbounds": [{
		"tag": "vmess",
		"protocol": "vmess",
		"port": {{ .VMessPort }},
		"settings": {
			"clients": []
		},
		"streamSettings": {
			"security": "tls",
			"tlsSettings": {
				"certificates": [{
					"certificateFile": "{{ .HomeDir }}/crt.crt",
					"keyFile": "{{ .HomeDir }}/key.key"
				}]
			}
		}
	}, {
		"tag": "trojan",
		"protocol": "trojan",
		"port": {{ .TrojanPort }},
		"settings": {
			"clients": []
		},
		"streamSettings": {
			"security": "tls",
			"tlsSettings": {
				"certificates": [{
					"certificateFile": "{{ .HomeDir }}/crt.crt",
					"keyFile": "{{ .HomeDir }}/key.key"
				}]
			}
		}
	}, {
		"tag": "api",
		"port": {{ .APIPort }},
		"protocol": "dokodemo-door",
		"settings": {
			"address": "127.0.0.1"
		}
	}],
	"outbounds": [{
		"protocol": "freedom"
	}],
	"routing": {
        "rules": [{
			"inboundTag": ["api"],
			"outboundTag": "api",
			"type": "field"
		}]
    }
}
`)
)
