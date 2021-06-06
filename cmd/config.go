package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sentinel-official/dvpn-node/types"
)

func ConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration sub-commands",
	}

	cmd.AddCommand(
		configInit(),
		configShow(),
	)

	return cmd
}

func configInit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the default configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				home = viper.GetString(flags.FlagHome)
				path = filepath.Join(home, types.ConfigFileName)
			)

			force, err := cmd.Flags().GetBool(types.FlagForce)
			if err != nil {
				return err
			}

			if !force {
				if _, err = os.Stat(path); err == nil {
					return fmt.Errorf("config file already exists at path %s", path)
				}
			}

			if err := os.MkdirAll(home, 0700); err != nil {
				return err
			}

			cfg := types.NewConfig().WithDefaultValues()
			return cfg.SaveToPath(path)
		},
	}

	cmd.Flags().Bool(types.FlagForce, false, "force")

	return cmd
}

func configShow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				home = viper.GetString(flags.FlagHome)
				path = filepath.Join(home, types.ConfigFileName)
			)

			cfg := types.NewConfig().WithDefaultValues()
			if err := cfg.LoadFromPath(path); err != nil {
				return err
			}

			fmt.Println(cfg.String())
			return nil
		},
	}

	return cmd
}
