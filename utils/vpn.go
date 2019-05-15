package utils

import (
	"github.com/pkg/errors"

	"github.com/ironman0x7b2/vpn-node/config"
	"github.com/ironman0x7b2/vpn-node/types"
	"github.com/ironman0x7b2/vpn-node/vpn/open_vpn"
)

func ProcessVPN(_type string) (types.BaseVPN, error) {
	switch _type {
	case "OpenVPN":
		return processOpenVPN()
	default:
		return nil, errors.Errorf("Currently the VPN type `%s` is not supported", _type)
	}
}

func processOpenVPN() (*open_vpn.OpenVPN, error) {
	cfg := config.NewOpenVPNConfig()

	if err := cfg.LoadFromPath(types.DefaultOpenVPNConfigFilePath); err != nil {
		return nil, err
	}

	defer func() {
		if err := cfg.SaveToPath(types.DefaultOpenVPNConfigFilePath); err != nil {
			panic(err)
		}
	}()

	publicIP, err := PublicIP()
	if err != nil {
		return nil, err
	}

	return open_vpn.NewOpenVPN(cfg.Port, publicIP, cfg.Protocol, cfg.EncryptionMethod), nil
}
