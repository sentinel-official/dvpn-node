package tx

import (
	"encoding/hex"
	"log"
	"path/filepath"
	"sync"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/pkg/errors"
	tmLog "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/lite/proxy"
	"github.com/tendermint/tendermint/rpc/client"
	core "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/sentinel-official/dvpn-node/cli"
	"github.com/sentinel-official/dvpn-node/types"
)

type Manager struct {
	CLI       *cli.CLI
	TxBuilder auth.TxBuilder
	password  string
	mutex     *sync.Mutex
}

// nolint: gocritic
func NewManager(cliContext *cli.CLI, _txBuilder auth.TxBuilder, password string) *Manager {
	return &Manager{
		CLI:       cliContext,
		TxBuilder: _txBuilder,
		password:  password,
		mutex:     &sync.Mutex{},
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

	_cli := cli.NewCLI(cdc, kb, rpcAddress, keyInfo)
	client, verifier, err := cli.NewVerifier(types.DefaultConfigDir, chainID, rpcAddress)
	if err != nil {
		return nil, err
	}

	_cli.Client = client
	_cli.Verifier = verifier

	account, err := _cli.GetAccount(keyInfo.GetAddress().Bytes())
	if err != nil {
		return nil, err
	}

	_txBuilder := auth.NewTxBuilder(utils.GetTxEncoder(cdc),
		account.GetAccountNumber(), account.GetSequence(), 1000000000,
		1.0, false, chainID,
		"", sdk.Coins{}, sdk.DecCoins{}).
		WithKeybase(kb)

	return NewManager(_cli, _txBuilder, password), nil
}

func (m *Manager) CompleteAndBroadcastTxSync(messages []sdk.Msg) (*sdk.TxResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	txBytes, err := m.TxBuilder.BuildAndSign(m.CLI.FromName, m.password, messages)
	if err != nil {
		return nil, err
	}

	node, err := m.CLI.GetNode()
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

	node, err := m.CLI.GetNode()
	if err != nil {
		return nil, err
	}

	log.Printf("Querying the transaction with hash `%s`", hash)
	res, err := node.Tx(_hash, !m.CLI.TrustNode)
	if err != nil {
		return nil, err
	}

	if !m.CLI.TrustNode {
		if err := utils.ValidateTxResult(m.CLI.CLIContext, res); err != nil {
			return nil, err
		}
	}

	return res, nil
}
