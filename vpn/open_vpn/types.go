package open_vpn

import (
	"path/filepath"
)

var (
	defaultOpenVPNDir           = "/etc/openvpn"
	defaultKeysDir              = filepath.Join(defaultOpenVPNDir, "keys")
	defaultServerConfigFilePath = filepath.Join(defaultOpenVPNDir, "server.conf")
	defaultClientConfigFilePath = filepath.Join(defaultOpenVPNDir, "client.conf")
	defaultStatusLogFilePath    = filepath.Join(defaultOpenVPNDir, "openvpn-status.log")
	defaultEasyRSADir           = "/usr/share/easy-rsa"
	cnamePrefix                 = "client_"
	ovpnFileExtension           = ".ovpn"
)

type serverConfigData struct {
	Port              uint32
	Protocol          string
	KeysDir           string
	Encryption        string
	StatusLogFilePath string
}

func newServerConfigData(port uint32, protocol, encryption string) serverConfigData {
	return serverConfigData{
		Port:              port,
		Protocol:          protocol,
		KeysDir:           defaultKeysDir,
		Encryption:        encryption,
		StatusLogFilePath: defaultStatusLogFilePath,
	}
}

type clientConfigData struct {
	IPAddress  string
	Port       uint32
	Protocol   string
	Encryption string
}

func newClientConfigData(ipAddress string, port uint32, protocol, encryption string) clientConfigData {
	return clientConfigData{
		IPAddress:  ipAddress,
		Port:       port,
		Protocol:   protocol,
		Encryption: encryption,
	}
}
