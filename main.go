package main

import (
	sent "github.com/sentinel-official/hub/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sentinel-official/dvpn-node/cmd"
	wireguard "github.com/sentinel-official/dvpn-node/services/wireguard/cli"
	"github.com/sentinel-official/dvpn-node/types"
)

func main() {
	sent.GetConfig().Seal()
	cobra.EnableCommandSorting = false

	root := &cobra.Command{
		Use:          "sentinel-dvpn-node",
		SilenceUsage: true,
	}

	root.AddCommand(
		cmd.ConfigCmd(),
		cmd.KeysCmd(),
		types.LineBreakCmd,
		wireguard.Command(),
		types.LineBreakCmd,
		cmd.StartCmd(),
	)

	root.PersistentFlags().String(types.FlagHome, types.DefaultHomeDirectory, "home")
	root.PersistentFlags().String(types.FlagLogLevel, "info", "log level")

	_ = viper.BindPFlag(types.FlagHome, root.PersistentFlags().Lookup(types.FlagHome))
	_ = viper.BindPFlag(types.FlagLogLevel, root.PersistentFlags().Lookup(types.FlagLogLevel))

	_ = root.Execute()
}
