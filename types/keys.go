package types

import (
	"os"
	"path/filepath"
)

const (
	ConfigFileName = "config.toml"
	FlagForce      = "force"
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
