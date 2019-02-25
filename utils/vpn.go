package utils

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/ironman0x7b2/vpn-node/config"
	"github.com/ironman0x7b2/vpn-node/types"
	"github.com/ironman0x7b2/vpn-node/vpn/open_vpn"
)

func ProcessVPN(vpnType string) (types.BaseVPN, error) {
	switch vpnType {
	case "OpenVPN":
		return processOpenVPN()
	default:
		return nil, errors.New(fmt.Sprintf("Invalid VPN type: %s", vpnType))
	}
}

func processOpenVPN() (*open_vpn.OpenVPN, error) {
	cfg := config.NewOpenVPNConfig()
	if err := cfg.LoadFromPath(""); err != nil {
		return nil, err
	}

	defer func() {
		if err := cfg.SaveToPath(""); err != nil {
			panic(err)
		}
	}()

	if len(cfg.PublicIP) == 0 {
		ip, err := PublicIP()
		if err != nil {
			return nil, err
		}

		cfg.PublicIP = ip
	}

	return open_vpn.NewOpenVPN(cfg.Port, cfg.ManagementPort, cfg.PublicIP, cfg.Protocol, cfg.Encryption), nil
}
