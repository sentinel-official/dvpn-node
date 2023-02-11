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
		Use:   "add (name)",
		Short: "Add a key",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				home       = viper.GetString(flags.FlagHome)
				configPath = filepath.Join(home, types.ConfigFileName)
			)

			v := viper.New()
			v.SetConfigFile(configPath)

			config, err := types.ReadInConfig(v)
			if err != nil {
				return err
			}

			skipConfigValidation, err := cmd.Flags().GetBool(flagSkipConfigValidation)
			if err != nil {
				return err
			}

			if !skipConfigValidation {
				if err := config.Validate(); err != nil {
					return err
				}
			}

			account, err := cmd.Flags().GetUint32(flagAccount)
			if err != nil {
				return err
			}

			index, err := cmd.Flags().GetUint32(flagIndex)
			if err != nil {
				return err
			}

			recoverKey, err := cmd.Flags().GetBool(flagRecover)
			if err != nil {
				return err
			}

			var (
				name   = config.Keyring.From
				reader = bufio.NewReader(cmd.InOrStdin())
			)

			if len(args) > 0 {
				name = args[0]
			}

			kr, err := keyring.New(sdk.KeyringServiceName(), config.Keyring.Backend, home, reader)
			if err != nil {
				return err
			}

			if _, err = kr.Key(name); err == nil {
				return fmt.Errorf("key already exists with name %s", name)
			}

			entropy, err := bip39.NewEntropy(256)
			if err != nil {
				return err
			}

			mnemonic, err := bip39.NewMnemonic(entropy)
			if err != nil {
				return err
			}

			if recoverKey {
				mnemonic, err = input.GetString("Enter your bip39 mnemonic", reader)
				if err != nil {
					return err
				}

				if !bip39.IsMnemonicValid(mnemonic) {
					return errors.New("invalid bip39 mnemonic")
				}
			}

			var (
				hdPath        = hd.CreateHDPath(sdk.GetConfig().GetCoinType(), account, index)
				algorithms, _ = kr.SupportedAlgorithms()
			)

			algorithm, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), algorithms)
			if err != nil {
				return err
			}

			key, err := kr.NewAccount(name, mnemonic, "", hdPath.String(), algorithm)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "operator: %s\n", key.GetAddress())
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "address: %s\n", hubtypes.NodeAddress(key.GetAddress()))
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "\n")
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "**Important** write this mnemonic phrase in a safe place\n")
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s\n", mnemonic)

			return nil
		},
	}

	cmd.Flags().Bool(flagSkipConfigValidation, false, "skip the validation of configuration")
	cmd.Flags().Bool(flagRecover, false, "provide mnemonic phrase to recover an existing key")
	cmd.Flags().Uint32(flagAccount, 0, "account number for HD derivation")
	cmd.Flags().Uint32(flagIndex, 0, "address index number for HD derivation")

	return cmd
}

func keysShow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show (name)",
		Short: "Show a key",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				home       = viper.GetString(flags.FlagHome)
				configPath = filepath.Join(home, types.ConfigFileName)
			)

			v := viper.New()
			v.SetConfigFile(configPath)

			config, err := types.ReadInConfig(v)
			if err != nil {
				return err
			}

			skipConfigValidation, err := cmd.Flags().GetBool(flagSkipConfigValidation)
			if err != nil {
				return err
			}

			if !skipConfigValidation {
				if err := config.Validate(); err != nil {
					return err
				}
			}

			var (
				name   = config.Keyring.From
				reader = bufio.NewReader(cmd.InOrStdin())
			)

			if len(args) > 0 {
				name = args[0]
			}

			kr, err := keyring.New(sdk.KeyringServiceName(), config.Keyring.Backend, home, reader)
			if err != nil {
				return err
			}

			key, err := kr.Key(name)
			if err != nil {
				return err
			}

			fmt.Printf("operator: %s\n", key.GetAddress())
			fmt.Printf("address: %s\n", hubtypes.NodeAddress(key.GetAddress()))

			return nil
		},
	}

	cmd.Flags().Bool(flagSkipConfigValidation, false, "skip the validation of configuration")

	return cmd
}

func keysList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all the keys",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var (
				home       = viper.GetString(flags.FlagHome)
				configPath = filepath.Join(home, types.ConfigFileName)
			)

			v := viper.New()
			v.SetConfigFile(configPath)

			config, err := types.ReadInConfig(v)
			if err != nil {
				return err
			}

			skipConfigValidation, err := cmd.Flags().GetBool(flagSkipConfigValidation)
			if err != nil {
				return err
			}

			if !skipConfigValidation {
				if err := config.Validate(); err != nil {
					return err
				}
			}

			var (
				reader = bufio.NewReader(cmd.InOrStdin())
			)

			kr, err := keyring.New(sdk.KeyringServiceName(), config.Keyring.Backend, home, reader)
			if err != nil {
				return err
			}

			keys, err := kr.List()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 1, 1, 1, ' ', 0)
			for _, key := range keys {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
					key.GetName(),
					key.GetAddress(),
					hubtypes.NodeAddress(key.GetAddress().Bytes()),
				)
			}

			return w.Flush()
		},
	}

	cmd.Flags().Bool(flagSkipConfigValidation, false, "skip the validation of configuration")

	return cmd
}

func keysDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete (name)",
		Short: "Delete a key",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				home       = viper.GetString(flags.FlagHome)
				configPath = filepath.Join(home, types.ConfigFileName)
			)

			v := viper.New()
			v.SetConfigFile(configPath)

			config, err := types.ReadInConfig(v)
			if err != nil {
				return err
			}

			skipConfigValidation, err := cmd.Flags().GetBool(flagSkipConfigValidation)
			if err != nil {
				return err
			}

			if !skipConfigValidation {
				if err := config.Validate(); err != nil {
					return err
				}
			}

			var (
				name   = config.Keyring.From
				reader = bufio.NewReader(cmd.InOrStdin())
			)

			if len(args) > 0 {
				name = args[0]
			}

			kr, err := keyring.New(sdk.KeyringServiceName(), config.Keyring.Backend, home, reader)
			if err != nil {
				return err
			}

			return kr.Delete(name)
		},
	}

	cmd.Flags().Bool(flagSkipConfigValidation, false, "skip the validation of configuration")

	return cmd
}
