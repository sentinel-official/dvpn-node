package types

import (
	"os"
	"path/filepath"
)

const (
	DefaultPassword = "0123456789"
	ConfigFileName  = "config.toml"
)

var (
	Version              = ""
	DefaultHomeDirectory = func() string {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		return filepath.Join(home, ".sentinel", "node")
	}()
)
