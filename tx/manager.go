package tx

import (
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	txBuilder "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/pkg/errors"
	tmLog "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/lite/proxy"
	"github.com/tendermint/tendermint/rpc/client"
	core "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/sentinel-official/sentinel-dvpn-node/types"
)

type Manager struct {
	CLIContext context.CLIContext
	TxBuilder  txBuilder.TxBuilder
	password   string
	mutex      *sync.Mutex
}

// nolint: gocritic
func NewManager(cliContext context.CLIContext, _txBuilder txBuilder.TxBuilder, password string) *Manager {
	return &Manager{
		CLIContext: cliContext,
		TxBuilder:  _txBuilder,
		password:   password,
		mutex:      &sync.Mutex{},
	}
}

func NewManagerWithConfig(chainID, rpcAddress, password string, cdc *codec.Codec, keyInfo keys.Info,
	kb keys.Keybase) (*Manager, error) {

	_client := client.NewHTTP(rpcAddress, "/websocket")

	verifier, err := proxy.NewVerifier(chainID, filepath.Join(types.DefaultConfigDir, "lite"),
		_client, tmLog.NewNopLogger(), 10)
	if err != nil {
		return nil, err
	}

	cliContext := context.CLIContext{
		Codec:         cdc,
		Client:        _client,
		Keybase:       kb,
		Output:        os.Stdout,
		OutputFormat:  "text",
		NodeURI:       rpcAddress,
		From:          keyInfo.GetName(),
		AccountStore:  auth.StoreKey,
		BroadcastMode: "sync",
		Verifier:      verifier,
		VerifierHome:  types.DefaultConfigDir,
		FromAddress:   keyInfo.GetAddress(),
		FromName:      keyInfo.GetName(),
		SkipConfirm:   true,
	}.WithAccountDecoder(cdc)

	account, err := cliContext.GetAccount(keyInfo.GetAddress().Bytes())
	if err != nil {
		return nil, err
	}

	_txBuilder := txBuilder.NewTxBuilder(utils.GetTxEncoder(cdc),
		account.GetAccountNumber(), account.GetSequence(), 1000000000,
		1.0, false, chainID,
		"", sdk.Coins{}, sdk.DecCoins{}).
		WithKeybase(kb)

	return NewManager(cliContext, _txBuilder, password), nil
}

func (m *Manager) CompleteAndBroadcastTxSync(messages []sdk.Msg) (*sdk.TxResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	txBytes, err := m.TxBuilder.BuildAndSign(m.CLIContext.FromName, m.password, messages)
	if err != nil {
		return nil, err
	}

	node, err := m.CLIContext.GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.BroadcastTxSync(txBytes)
	if res != nil && res.Code == 0 {
		m.TxBuilder = m.TxBuilder.WithSequence(m.TxBuilder.Sequence() + 1)
	}

	txRes := sdk.NewResponseFormatBroadcastTx(res)
	if txRes.Code != 0 {
		return &txRes, errors.Errorf(txRes.String())
	}

	return &txRes, err
}

func (m *Manager) QueryTx(hash string) (*core.ResultTx, error) {
	_hash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	node, err := m.CLIContext.GetNode()
	if err != nil {
		return nil, err
	}

	log.Printf("Querying the transaction with hash `%s`", hash)
	res, err := node.Tx(_hash, !m.CLIContext.TrustNode)
	if err != nil {
		return nil, err
	}

	if !m.CLIContext.TrustNode {
		if err := tx.ValidateTxResult(m.CLIContext, res); err != nil {
			return nil, err
		}
	}

	return res, nil
}
