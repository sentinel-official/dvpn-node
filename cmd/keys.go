package cmd

import (
	"bufio"
	"fmt"
	"path/filepath"
	"text/tabwriter"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
	hubtypes "github.com/sentinel-official/hub/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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
			var (
				home = viper.GetString(flags.FlagHome)
				path = filepath.Join(home, types.ConfigFileName)
			)

			v := viper.New()
			v.SetConfigFile(path)

			cfg, err := types.ReadInConfig(v)
			if err != nil {
				return err
			}

			skipConfigValidation, err := cmd.Flags().GetBool(flagSkipConfigValidation)
			if err != nil {
				return err
			}

			if !skipConfigValidation {
				if err := cfg.Validate(); err != nil {
					return err
				}
			}

			recovery, err := cmd.Flags().GetBool(flagRecover)
			if err != nil {
				return err
			}

			var (
				reader = bufio.NewReader(cmd.InOrStdin())
			)

			kr, err := keyring.New(sdk.KeyringServiceName(), cfg.Keyring.Backend, home, reader)
			if err != nil {
				return err
			}

			if _, err = kr.Key(args[0]); err == nil {
				return fmt.Errorf("key already exists with name %s", args[0])
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
				mnemonic, err = input.GetString("Enter your bip39 mnemonic", reader)
				if err != nil {
					return err
				}

				if !bip39.IsMnemonicValid(mnemonic) {
					return errors.New("invalid bip39 mnemonic")
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

			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "operator: %s\n", info.GetAddress())
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "address: %s\n", hubtypes.NodeAddress(info.GetAddress().Bytes()))
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "")
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "**Important** write this mnemonic phrase in a safe place")
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), mnemonic)

			return nil
		},
	}

	cmd.Flags().Bool(flagRecover, false, "recover")
	cmd.Flags().Bool(flagSkipConfigValidation, false, "skip the validation of configuration file")

	return cmd
}

func keysShow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [name]",
		Short: "Show a key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				home = viper.GetString(flags.FlagHome)
				path = filepath.Join(home, types.ConfigFileName)
			)

			v := viper.New()
			v.SetConfigFile(path)

			cfg, err := types.ReadInConfig(v)
			if err != nil {
				return err
			}

			skipConfigValidation, err := cmd.Flags().GetBool(flagSkipConfigValidation)
			if err != nil {
				return err
			}

			if !skipConfigValidation {
				if err := cfg.Validate(); err != nil {
					return err
				}
			}

			var (
				reader = bufio.NewReader(cmd.InOrStdin())
			)

			kr, err := keyring.New(sdk.KeyringServiceName(), cfg.Keyring.Backend, home, reader)
			if err != nil {
				return err
			}

			info, err := kr.Key(args[0])
			if err != nil {
				return err
			}

			fmt.Printf("operator: %s\n", info.GetAddress())
			fmt.Printf("address: %s\n", hubtypes.NodeAddress(info.GetAddress().Bytes()))

			return nil
		},
	}

	cmd.Flags().Bool(flagSkipConfigValidation, false, "skip the validation of configuration file")

	return cmd
}

func keysList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all the keys",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var (
				home = viper.GetString(flags.FlagHome)
				path = filepath.Join(home, types.ConfigFileName)
			)

			v := viper.New()
			v.SetConfigFile(path)

			cfg, err := types.ReadInConfig(v)
			if err != nil {
				return err
			}

			skipConfigValidation, err := cmd.Flags().GetBool(flagSkipConfigValidation)
			if err != nil {
				return err
			}

			if !skipConfigValidation {
				if err := cfg.Validate(); err != nil {
					return err
				}
			}

			var (
				reader = bufio.NewReader(cmd.InOrStdin())
			)

			kr, err := keyring.New(sdk.KeyringServiceName(), cfg.Keyring.Backend, home, reader)
			if err != nil {
				return err
			}

			infos, err := kr.List()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 1, 1, 1, ' ', 0)
			for _, info := range infos {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
					info.GetName(),
					info.GetAddress(),
					hubtypes.NodeAddress(info.GetAddress().Bytes()),
				)
			}

			return w.Flush()
		},
	}

	cmd.Flags().Bool(flagSkipConfigValidation, false, "skip the validation of configuration file")

	return cmd
}

func keysDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				home = viper.GetString(flags.FlagHome)
				path = filepath.Join(home, types.ConfigFileName)
			)

			v := viper.New()
			v.SetConfigFile(path)

			cfg, err := types.ReadInConfig(v)
			if err != nil {
				return err
			}

			skipConfigValidation, err := cmd.Flags().GetBool(flagSkipConfigValidation)
			if err != nil {
				return err
			}

			if !skipConfigValidation {
				if err := cfg.Validate(); err != nil {
					return err
				}
			}

			var (
				reader = bufio.NewReader(cmd.InOrStdin())
			)

			kr, err := keyring.New(sdk.KeyringServiceName(), cfg.Keyring.Backend, home, reader)
			if err != nil {
				return err
			}

			return kr.Delete(args[0])
		},
	}

	cmd.Flags().Bool(flagSkipConfigValidation, false, "skip the validation of configuration file")

	return cmd
}
