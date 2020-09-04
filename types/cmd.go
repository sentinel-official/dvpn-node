package types

import (
	"github.com/spf13/cobra"
)

const (
	FlagHome     = "home"
	FlagForce    = "force"
	FlagLogLevel = "log-level"
)

var (
	LineBreakCmd = &cobra.Command{Run: func(_ *cobra.Command, _ []string) {}}
)
