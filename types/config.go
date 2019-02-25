package types

import (
	"os"
	"path/filepath"
	"time"
)

var (
	HomeDir                      = os.ExpandEnv("$HOME")
	DefaultConfigDir             = filepath.Join(HomeDir, ".sentinel")
	DefaultAppConfigFilePath     = filepath.Join(DefaultConfigDir, "app_config.json")
	DefaultOpenVPNConfigFilePath = filepath.Join(DefaultConfigDir, "open_vpn_config.json")
	Version                      = "0.2.0"

	IntervalUpdateNodeStatus        = 200 * time.Second
	IntervalUpdateSessionsBandwidth = 100 * time.Second
	IntervalRequestBandwidthSigns   = 5 * time.Second
	TimeoutConnectionRead           = 30 * time.Second
	TimeoutSession                  = 30 * time.Second
)
