package utils

import (
	"net"

	"github.com/pkg/errors"

	"github.com/sentinel-official/dvpn-node/config"
	"github.com/sentinel-official/dvpn-node/types"
	openvpn "github.com/sentinel-official/dvpn-node/vpn/open_vpn"
)

func ProcessVPN(_type string, ip net.IP) (types.BaseVPN, error) {
	switch _type {
	case openvpn.Type:
		return processOpenVPN(ip)
	default:
		return nil, errors.Errorf("VPN `%s` is not supported", _type)
	}
}

func processOpenVPN(ip net.IP) (*openvpn.OpenVPN, error) {
	cfg := config.NewOpenVPNConfig()
	if err := cfg.LoadFromPath(types.DefaultOpenVPNConfigFilePath); err != nil {
		return nil, err
	}

	defer func() {
		if err := cfg.SaveToPath(types.DefaultOpenVPNConfigFilePath); err != nil {
			panic(err)
		}
	}()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return openvpn.NewOpenVPN(cfg.Port, ip, cfg.Protocol, cfg.Encryption), nil
}
