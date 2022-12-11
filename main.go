package main

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	hubtypes "github.com/sentinel-official/hub/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sentinel-official/dvpn-node/cmd"
	wireguard "github.com/sentinel-official/dvpn-node/services/wireguard/cli"
	"github.com/sentinel-official/dvpn-node/types"
)

func main() {
	hubtypes.GetConfig().Seal()
	root := &cobra.Command{
		Use:          "sentinelnode",
		SilenceUsage: true,
	}

	root.AddCommand(
		cmd.ConfigCmd(),
		cmd.KeysCmd(),
		wireguard.Command(),
		cmd.StartCmd(),
		version.NewVersionCommand(),
	)

	root.PersistentFlags().String(flags.FlagHome, types.DefaultHomeDirectory, "home")
	root.PersistentFlags().String(flags.FlagLogFormat, "plain", "log format")
	root.PersistentFlags().String(flags.FlagLogLevel, "info", "log level")

	_ = viper.BindPFlag(flags.FlagHome, root.PersistentFlags().Lookup(flags.FlagHome))
	_ = viper.BindPFlag(flags.FlagLogFormat, root.PersistentFlags().Lookup(flags.FlagLogFormat))
	_ = viper.BindPFlag(flags.FlagLogLevel, root.PersistentFlags().Lookup(flags.FlagLogLevel))

	_ = root.Execute()
}
