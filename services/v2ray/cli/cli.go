package cli

import (
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "v2ray",
		Aliases: nil,
		Short:   "V2Ray sub-commands",
	}

	cmd.AddCommand(
		configCmd(),
	)

	return cmd
}
