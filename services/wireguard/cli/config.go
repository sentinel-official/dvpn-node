package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	wgt "github.com/sentinel-official/dvpn-node/services/wireguard/types"
	"github.com/sentinel-official/dvpn-node/types"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Config",
	}

	cmd.AddCommand(
		configInitCmd(),
		configShowCmd(),
	)

	return cmd
}

func configInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Init",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			force, err := cmd.Flags().GetBool(types.FlagForce)
			if err != nil {
				return err
			}

			configPath := filepath.Join(home, wgt.ConfigFileName)

			if !force {
				_, err = os.Stat(configPath)
				if err == nil {
					return fmt.Errorf("config file already exists at path '%s'", configPath)
				}
			}

			if err := os.MkdirAll(home, 0700); err != nil {
				return err
			}

			config := wgt.NewConfig().WithDefaultValues()
			return config.SaveToPath(configPath)
		},
	}

	cmd.Flags().Bool(types.FlagForce, false, "force")

	return cmd
}

func configShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			config := wgt.NewConfig()
			if err := config.LoadFromPath(filepath.Join(home, wgt.ConfigFileName)); err != nil {
				return err
			}

			fmt.Println(config)
			return nil
		},
	}

	return cmd
}
