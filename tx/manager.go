package tx

import (
	"encoding/hex"
	"log"
	"path/filepath"
	"sync"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	txBuilder "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/pkg/errors"
	tmLog "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/lite/proxy"
	"github.com/tendermint/tendermint/rpc/client"
	coreTypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/ironman0x7b2/vpn-node/config"
	"github.com/ironman0x7b2/vpn-node/types"
)

type Manager struct {
	CLIContext context.CLIContext
	TxBuilder  txBuilder.TxBuilder
	password   string
	mutex      *sync.Mutex
}

func NewManager(cliContext context.CLIContext, txBuilder txBuilder.TxBuilder, password string) *Manager {
	return &Manager{
		CLIContext: cliContext,
		TxBuilder:  txBuilder,
		password:   password,
		mutex:      &sync.Mutex{},
	}
}

func NewManagerFromConfig(appCfg *config.AppConfig, cdc *codec.Codec, info keys.Info, kb keys.Keybase) (*Manager, error) {
	verifier, err := proxy.NewVerifier(appCfg.ChainID,
		filepath.Join(types.DefaultConfigDir, ".vpn_lite"),
		client.NewHTTP(appCfg.RPCAddress, "/websocket"),
		tmLog.NewNopLogger(), 10)
	if err != nil {
		return nil, err
	}

	cliContext := context.NewCLIContext().
		WithCodec(cdc).
		WithAccountDecoder(cdc).
		WithNodeURI(appCfg.RPCAddress).
		WithVerifier(verifier).
		WithFrom(info.GetName()).
		WithFromName(info.GetName()).
		WithFromAddress(info.GetAddress())

	account, err := cliContext.GetAccount(info.GetAddress().Bytes())
	if err != nil {
		return nil, err
	}

	_txBuilder := txBuilder.NewTxBuilder(utils.GetTxEncoder(cdc),
		account.GetAccountNumber(), account.GetSequence(), 1000000000,
		1.0, false, appCfg.ChainID,
		"", csdkTypes.Coins{}, csdkTypes.DecCoins{}).WithKeybase(kb)

	return NewManager(cliContext, _txBuilder, appCfg.Account.Password), nil
}

func (m *Manager) CompleteAndBroadcastTxSync(messages []csdkTypes.Msg) (*csdkTypes.TxResponse, error) {
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

	txRes := csdkTypes.NewResponseFormatBroadcastTx(res)
	if txRes.Code != 0 {
		return &txRes, errors.Errorf(txRes.String())
	}

	return &txRes, err
}

func (m *Manager) QueryTx(hash string) (*coreTypes.ResultTx, error) {
	_hash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	node, err := m.CLIContext.GetNode()
	if err != nil {
		return nil, err
	}

	log.Printf("Querying the transaction details with hash `%s`", hash)
	res, err := node.Tx(_hash, !m.CLIContext.TrustNode)
	if err != nil {
		return nil, err
	}

	if !m.CLIContext.TrustNode {
		log.Println("Validating the queried transaction details")
		if err := tx.ValidateTxResult(m.CLIContext, res); err != nil {
			return nil, err
		}
	}

	return res, nil
}
