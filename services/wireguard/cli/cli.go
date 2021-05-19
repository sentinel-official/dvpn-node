package cli

import (
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wireguard",
		Short: "WireGuard sub-commands",
	}

	cmd.AddCommand(
		configCmd(),
	)

	return cmd
}
