package config

import (
	"os"
	"path/filepath"
)

var (
	HomeDir                  = os.ExpandEnv("$HOME")
	DefaultConfigDir         = filepath.Join(HomeDir, ".sentinel")
	DefaultAppConfigFilePath = filepath.Join(DefaultConfigDir, "app_config.json")
	Version                  = ""
)
