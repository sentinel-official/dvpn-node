package types

import (
	"os"
	"path/filepath"
)

const (
	ConfigFileName   = "config.toml"
	DatabaseFileName = "data.db"
	KeyringName      = "sentinel"
)

const (
	FlagForce = "force"
)

var (
	DefaultHomeDirectory = func() string {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		return filepath.Join(home, ".sentinelnode")
	}()
)
