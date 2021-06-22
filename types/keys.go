package types

import (
	"os"
	"path/filepath"
)

const (
	ConfigFileName  = "config.toml"
	FlagForce       = "force"
	KeyringName     = "sentinel"
	DefaultIPv4CIDR = "10.8.0.2/24"
	DefaultIPv6CIDR = "fd86:ea04:1115::2/120"
)

var (
	DefaultHomeDirectory = func() string {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		return filepath.Join(home, ".sentinel", "node")
	}()
)
