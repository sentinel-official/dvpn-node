package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	hubtypes "github.com/sentinel-official/hub/types"
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

			var (
				reader = bufio.NewReader(cmd.InOrStdin())
			)

			kr, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendFile, home, reader)
			if err != nil {
				return err
			}

			if _, err = kr.Key(args[0]); err == nil {
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

			var (
				hdPath                 = hd.CreateHDPath(sdk.GetConfig().GetCoinType(), 0, 0)
				supportedAlgorithms, _ = kr.SupportedAlgorithms()
			)

			signingAlgorithm, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), supportedAlgorithms)
			if err != nil {
				return err
			}

			info, err := kr.NewAccount(args[0], mnemonic, "", hdPath.String(), signingAlgorithm)
			if err != nil {
				return err
			}

			fmt.Printf("Address:  %s\n", hubtypes.NodeAddress(info.GetAddress().Bytes()))
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

			var (
				reader = bufio.NewReader(cmd.InOrStdin())
			)

			kr, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendFile, home, reader)
			if err != nil {
				return err
			}

			info, err := kr.Key(args[0])
			if err != nil {
				return err
			}

			fmt.Printf("Address:  %s\n", hubtypes.NodeAddress(info.GetAddress().Bytes()))
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

			var (
				reader = bufio.NewReader(cmd.InOrStdin())
			)

			kr, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendFile, home, reader)
			if err != nil {
				return err
			}

			list, err := kr.List()
			if err != nil {
				return err
			}

			for _, info := range list {
				fmt.Printf("%s | %s | %s\n",
					info.GetName(), hubtypes.NodeAddress(info.GetAddress().Bytes()), info.GetAddress())
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

			var (
				reader = bufio.NewReader(cmd.InOrStdin())
			)

			kr, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendFile, home, reader)
			if err != nil {
				return err
			}

			return kr.Delete(args[0])
		},
	}

	return cmd
}
