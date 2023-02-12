package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	v2raytypes "github.com/sentinel-official/dvpn-node/services/v2ray/types"
	"github.com/sentinel-official/dvpn-node/types"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration sub-commands",
	}

	cmd.AddCommand(
		configInit(),
		configShow(),
		configSet(),
	)

	return cmd
}

func configInit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Init the configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				home = viper.GetString(flags.FlagHome)
				path = filepath.Join(home, v2raytypes.ConfigFileName)
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

			config := v2raytypes.NewConfig().WithDefaultValues()
			return config.SaveToPath(path)
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
				path = filepath.Join(home, v2raytypes.ConfigFileName)
			)

			v := viper.New()
			v.SetConfigFile(path)

			cfg, err := v2raytypes.ReadInConfig(v)
			if err != nil {
				return err
			}

			fmt.Println(cfg.String())
			return nil
		},
	}

	return cmd
}

func configSet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set the configuration",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			var (
				home = viper.GetString(flags.FlagHome)
				path = filepath.Join(home, v2raytypes.ConfigFileName)
			)

			v := viper.New()
			v.SetConfigFile(path)

			cfg, err := v2raytypes.ReadInConfig(v)
			if err != nil {
				return err
			}

			v.Set(args[0], args[1])

			if err := v.Unmarshal(cfg); err != nil {
				return err
			}

			return cfg.SaveToPath(path)
		},
	}

	return cmd
}
