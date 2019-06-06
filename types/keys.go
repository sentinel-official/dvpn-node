package types

import (
	"os"
	"path/filepath"
	"time"
)

const (
	INIT     = "INIT"
	ACTIVE   = "ACTIVE"
	INACTIVE = "INACTIVE"
)

// nolint:gochecknoglobals
var (
	Version                            = ""
	HomeDir                            = os.ExpandEnv("$HOME")
	DefaultConfigDir                   = filepath.Join(HomeDir, ".sentinel", "vpn-node")
	DefaultAppConfigFilePath           = filepath.Join(DefaultConfigDir, "app_config.toml")
	DefaultOpenVPNConfigFilePath       = filepath.Join(DefaultConfigDir, "open_vpn_config.toml")
	ConnectionReadTimeout              = 30 * time.Second
	SessionTimeout                     = 30 * time.Second
	RequestBandwidthSignInterval       = 5 * time.Second
	UpdateNodeStatusInterval           = 200 * time.Second
	UpdateSessionBandwidthInfoInterval = 100 * time.Second
)
