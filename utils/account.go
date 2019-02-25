package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	clientKeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
)

const mnemonicEntropySize = 256

func createAccount(keybase keys.Keybase) (keys.Info, error) {
	var name string

	fmt.Printf("Enter a new account name: ")
	name, err := client.BufferStdin().ReadString('\n')
	if err != nil {
		return nil, err
	}

	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return nil, errors.New("Entered account name is empty")
	}

	if _, err := keybase.Get(name); err == nil {
		return nil, errors.New(fmt.Sprintf("Account already exists with name `%s`", name))
	}

	password, err := client.GetCheckPassword(
		"Enter a passphrase to encrypt your key to disk: ",
		"Repeat the passphrase: ", client.BufferStdin())
	if err != nil {
		return nil, err
	}

	prompt := "Enter your bip39 mnemonic, or hit enter to generate one."
	mnemonic, err := client.GetString(prompt, client.BufferStdin())
	if err != nil {
		return nil, err
	}

	if len(mnemonic) == 0 {
		entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
		if err != nil {
			return nil, err
		}

		mnemonic, err = bip39.NewMnemonic(entropySeed)
		if err != nil {
			return nil, err
		}
	}

	info, err := keybase.CreateAccount(name, mnemonic, keys.DefaultBIP39Passphrase, password, uint32(0), uint32(0))
	if err != nil {
		return nil, err
	}

	_, _ = fmt.Fprintf(os.Stderr, "**Important** write this mnemonic phrase in a safe place.\n"+
		"It is the only way to recover your account if you ever forget your password.\n\n"+
		"%s\n\n", mnemonic)

	return info, err
}

func ProcessOwnerAccount(keybase keys.Keybase, name string) (keys.Info, error) {
	if len(name) > 0 {
		return keybase.Get(name)
	}

	infos, err := keybase.List()
	if err != nil {
		return nil, err
	}
	if len(infos) == 0 {
		return createAccount(keybase)
	}

	accounts, err := clientKeys.Bech32KeysOutput(infos)
	if err != nil {
		return nil, err
	}

	fmt.Printf("NAME:\tTYPE:\tADDRESS:\t\t\t\t\t\tPUBKEY:\n")
	for _, account := range accounts {
		fmt.Printf("%s\t%s\t%s\t%s\n", account.Name, account.Type, account.Address, account.PubKey)
	}

	prompt := "Enter the account name from above list, or hit enter to create a new account."
	name, err = client.GetString(prompt, client.BufferStdin())
	if err != nil {
		return nil, err
	}
	if len(name) == 0 {
		return createAccount(keybase)
	}

	return keybase.Get(name)
}

func ProcessAccountPassword(keybase keys.Keybase, name string) (string, error) {
	promt := fmt.Sprintf("Enter the password of the account `%s`: ", name)
	password, err := client.GetPassword(promt, client.BufferStdin())
	if err != nil {
		return "", err
	}

	password = strings.TrimSpace(password)

	_, _, err = keybase.Sign(name, password, []byte(""))
	if err != nil {
		return "", err
	}

	return password, nil
}
