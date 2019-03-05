package tx

import (
	"encoding/hex"
	"errors"
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
	ownerInfo keys.Info, keybase keys.Keybase) (*Manager, error) {

	log.Println("Initializing the chain verifier")
	verifier, err := proxy.NewVerifier(appCfg.ChainID, filepath.Join(types.DefaultConfigDir, ".vpn_lite"),
		client.NewHTTP(appCfg.LiteClientURI, "/websocket"), tmLog.NewNopLogger(), 10)
	if err != nil {
		return nil, err
	}

	log.Println("Initializing the CLI context")
	cliContext := context.NewCLIContext().
		WithCodec(cdc).
		WithAccountDecoder(cdc).
		WithNodeURI(appCfg.LiteClientURI).
		WithVerifier(verifier).
		WithFrom(ownerInfo.GetName()).
		WithFromName(ownerInfo.GetName()).
		WithFromAddress(ownerInfo.GetAddress())

	log.Println("Fetching the owner account info")
	account, err := cliContext.GetAccount(ownerInfo.GetAddress().Bytes())
	if err != nil {
		return nil, err
	}

	log.Println("Initializing the transaction builder")
	txBuilder := clientTxBuilder.NewTxBuilderFromCLI().
		WithKeybase(keybase).
		WithAccountNumber(account.GetAccountNumber()).
		WithSequence(account.GetSequence()).
		WithChainID(appCfg.ChainID).
		WithGas(1000000000).
		WithTxEncoder(clientUtils.GetTxEncoder(cdc))

	return NewManager(cliContext, txBuilder, appCfg.Owner.Password), nil
}

func (m *Manager) CompleteAndBroadcastTxSync(msgs []csdkTypes.Msg) (*csdkTypes.TxResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	txBytes, err := m.TxBuilder.BuildAndSign(m.CLIContext.GetFromName(), m.password, msgs)
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
		return &txRes, errors.New(txRes.String())
	}

	return &txRes, err
}

func (m *Manager) QueryTx(txHash string) (*coreTypes.ResultTx, error) {
	txHashBytes, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, err
	}

	node, err := m.CLIContext.GetNode()
	if err != nil {
		return nil, err
	}

	log.Printf("Querying the transaction with hash `%s`", txHash)
	res, err := node.Tx(txHashBytes, !m.CLIContext.TrustNode)
	if err != nil {
		return nil, err
	}

	if !m.CLIContext.TrustNode {
		log.Println("Validating the query transaction response")
		if err := tx.ValidateTxResult(m.CLIContext, res); err != nil {
			return nil, err
		}
	}

	return res, nil
}
