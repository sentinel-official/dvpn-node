package types

import (
	"os"
	"path/filepath"
)

const (
	ConfigFileName   = "config.toml"
	DatabaseFileName = "data.db"
	IPv4CIDR         = "10.8.0.2/24"
	IPv6CIDR         = "fd86:ea04:1115::2/120"
	KeyringName      = "sentinel"
)

const (
	FlagForce = "force"
)

var DefaultHomeDirectory = func() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(home, ".sentinelnode")
}()
