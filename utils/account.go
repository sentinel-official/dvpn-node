package utils

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/go-bip39"
)

const mnemonicEntropySize = 256

func CreateAccount(keybase keys.Keybase, name, password string) (string, keys.Info, error) {
	entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
	if err != nil {
		return "", nil, err
	}

	mnemonic, err := bip39.NewMnemonic(entropySeed[:])
	if err != nil {
		return "", nil, err
	}

	if _, err := keybase.Get(name); err == nil {
		return "", nil, fmt.Errorf("account already exists with name")
	}

	account, err := keybase.CreateAccount(name, mnemonic, keys.DefaultBIP39Passphrase, password, uint32(0), uint32(0))
	if err != nil {
		return "", nil, err
	}

	return mnemonic, account, nil
}
