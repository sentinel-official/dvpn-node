package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

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
		Short: "Initialize the configuration with default values",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			force, err := cmd.Flags().GetBool(types.FlagForce)
			if err != nil {
				return err
			}

			path := filepath.Join(home, "config.toml")

			if !force {
				_, err = os.Stat(path)
				if err == nil {
					return fmt.Errorf("config file already exists at path '%s'", path)
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
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			path := filepath.Join(home, "config.toml")
			if _, err := os.Stat(path); err != nil {
				return fmt.Errorf("config file does not exist at path '%s'", path)
			}

			cfg := types.NewConfig()
			if err := cfg.LoadFromPath(path); err != nil {
				return err
			}

			fmt.Println(cfg)
			return nil
		},
	}

	return cmd
}
