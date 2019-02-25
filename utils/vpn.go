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
		cfg := config.NewOpenVPNConfig()
		if err := cfg.LoadFromPath(""); err != nil {
			return nil, err
		}

		return open_vpn.NewOpenVPN(cfg.Port, cfg.ManagementPort, cfg.IPAddress, cfg.Protocol, cfg.EncryptionMethod), nil
	default:
		return nil, errors.New("Invalid config: vpn_type")
	}
}
