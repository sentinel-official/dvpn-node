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
	Version                      = ""
	HomeDir                      = os.ExpandEnv("$HOME")
	DefaultConfigDir             = filepath.Join(HomeDir, ".sentinel", "vpn_node")
	DefaultDatabaseFilePath      = filepath.Join(DefaultConfigDir, "database.db")
	DefaultAppConfigFilePath     = filepath.Join(DefaultConfigDir, "app_config.toml")
	DefaultOpenVPNConfigFilePath = filepath.Join(DefaultConfigDir, "open_vpn_config.toml")
	DefaultTLSCertFilePath       = filepath.Join(DefaultConfigDir, "server.crt")
	DefaultTLSKeyFilePath        = filepath.Join(DefaultConfigDir, "server.key")
	ConnectionReadTimeout        = 30000000 * time.Second
	RequestBandwidthSignInterval = 500000 * time.Second
	UpdateBandwidthInfosInterval = 100000000 * time.Second
)
