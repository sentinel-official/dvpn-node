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

	RequestBandwidthSignInterval       = 5 * time.Second
	ConnectionReadTimeout              = 30 * time.Second
	SessionTimeout                     = 30 * time.Second
	UpdateNodeStatusInterval           = 200 * time.Second
	UpdateSessionBandwidthInfoInterval = 100 * time.Second
)
