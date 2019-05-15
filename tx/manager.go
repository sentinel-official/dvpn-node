package tx

import (
	"encoding/hex"
	"log"
	"path/filepath"
	"sync"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	clientUtils "github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	clientTxBuilder "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
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
	TxBuilder  clientTxBuilder.TxBuilder
	password   string
	mutex      sync.Mutex
}

func NewManager(cliContext context.CLIContext, txBuilder clientTxBuilder.TxBuilder, password string) *Manager {
	return &Manager{
		CLIContext: cliContext,
		TxBuilder:  txBuilder,
		password:   password,
		mutex:      sync.Mutex{},
	}
}

func NewManagerFromConfig(appCfg *config.AppConfig, cdc *codec.Codec,
	ownerInfo keys.Info, kb keys.Keybase) (*Manager, error) {

	verifier, err := proxy.NewVerifier(appCfg.ChainID,
		filepath.Join(types.DefaultConfigDir, ".vpn_lite"),
		client.NewHTTP(appCfg.RPCServerAddress, "/websocket"),
		tmLog.NewNopLogger(), 10)
	if err != nil {
		return nil, err
	}

	cliContext := context.NewCLIContext().
		WithCodec(cdc).
		WithAccountDecoder(cdc).
		WithNodeURI(appCfg.RPCServerAddress).
		WithVerifier(verifier).
		WithFrom(ownerInfo.GetName()).
		WithFromName(ownerInfo.GetName()).
		WithFromAddress(ownerInfo.GetAddress())

	account, err := cliContext.GetAccount(ownerInfo.GetAddress().Bytes())
	if err != nil {
		return nil, err
	}

	txBuilder := clientTxBuilder.NewTxBuilder(clientUtils.GetTxEncoder(cdc),
		account.GetAccountNumber(), account.GetSequence(), 1000000000,
		1.0, false, appCfg.ChainID,
		"", csdkTypes.Coins{}, csdkTypes.DecCoins{}).WithKeybase(kb)

	return NewManager(cliContext, txBuilder, appCfg.Account.Password), nil
}

func (m *Manager) CompleteAndBroadcastTxSync(messages []csdkTypes.Msg) (*csdkTypes.TxResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	txBytes, err := m.TxBuilder.BuildAndSign(m.CLIContext.GetFromName(), m.password, messages)
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
	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	node, err := m.CLIContext.GetNode()
	if err != nil {
		return nil, err
	}

	log.Printf("Querying the transaction details with hash `%s`", hash)
	res, err := node.Tx(hashBytes, !m.CLIContext.TrustNode)
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
