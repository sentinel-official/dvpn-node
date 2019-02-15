package tx

import (
	"encoding/hex"
	"sync"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	clientTxBuilder "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/tendermint/tendermint/rpc/core/types"
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

	return &txRes, err
}

func (m *Manager) QueryTx(txHash string) (*core_types.ResultTx, error) {
	txHashBytes, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, err
	}

	node, err := m.CLIContext.GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.Tx(txHashBytes, !m.CLIContext.TrustNode)
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
