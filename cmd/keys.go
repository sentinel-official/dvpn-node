package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/client/keys"
	ckeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/go-bip39"
	hub "github.com/sentinel-official/hub/types"
	"github.com/spf13/cobra"

	"github.com/sentinel-official/dvpn-node/types"
)

func KeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Keys sub-commands",
	}

	cmd.AddCommand(
		keysAdd(),
		keysShow(),
		keysList(),
		keysDelete(),
	)

	return cmd
}

func keysAdd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [name]",
		Short: "Add a key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			recovery, err := cmd.Flags().GetBool(flagRecover)
			if err != nil {
				return err
			}

			kb, err := keys.NewKeyBaseFromDir(home)
			if err != nil {
				return err
			}

			if _, err = kb.Get(args[0]); err == nil {
				return fmt.Errorf("key already exists with name '%s'", args[0])
			}

			entropy, err := bip39.NewEntropy(256)
			if err != nil {
				return err
			}

			mnemonic, err := bip39.NewMnemonic(entropy)
			if err != nil {
				return err
			}

			if recovery {
				mnemonic, err = input.GetString("Enter your bip39 mnemonic.", bufio.NewReader(os.Stdin))
				if err != nil {
					return err
				}

				if !bip39.IsMnemonicValid(mnemonic) {
					return fmt.Errorf("invalid bip39 mnemonic")
				}
			}

			info, err := kb.CreateAccount(args[0], mnemonic, "", types.DefaultPassword,
				ckeys.CreateHDPath(0, 0).String(), ckeys.Secp256k1)
			if err != nil {
				return err
			}

			fmt.Printf("Address:  %s\n", hub.NodeAddress(info.GetAddress().Bytes()))
			fmt.Printf("Operator: %s\n", info.GetAddress())
			fmt.Printf("Mnemonic: %s\n", mnemonic)

			return nil
		},
	}

	cmd.Flags().Bool(flagRecover, false, "recover")

	return cmd
}

func keysShow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [name]",
		Short: "Show a key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			kb, err := keys.NewKeyBaseFromDir(home)
			if err != nil {
				return err
			}

			info, err := kb.Get(args[0])
			if err != nil {
				return err
			}

			fmt.Printf("Address:  %s\n", hub.NodeAddress(info.GetAddress().Bytes()))
			fmt.Printf("Operator: %s\n", info.GetAddress())

			return nil
		},
	}

	return cmd
}

func keysList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all the keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			kb, err := keys.NewKeyBaseFromDir(home)
			if err != nil {
				return err
			}

			list, err := kb.List()
			if err != nil {
				return err
			}

			for _, info := range list {
				fmt.Printf("%s | %s | %s\n",
					info.GetName(), hub.NodeAddress(info.GetAddress().Bytes()), info.GetAddress())
			}

			return nil
		},
	}

	return cmd
}

func keysDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			kb, err := keys.NewKeyBaseFromDir(home)
			if err != nil {
				return err
			}

			return kb.Delete(args[0], "", true)
		},
	}

	return cmd
}
