package tx

import (
	"sync"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types"
	clientTxBuilder "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
)

type Manager struct {
	CLIContext *context.CLIContext
	TxBuilder  clientTxBuilder.TxBuilder
	password   string
	mutex      sync.Mutex
}

func NewManager(cliContext *context.CLIContext, txBuilder clientTxBuilder.TxBuilder, password string) *Manager {
	return &Manager{
		CLIContext: cliContext,
		TxBuilder:  txBuilder,
		password:   password,
		mutex:      sync.Mutex{},
	}
}

func (m *Manager) CompleteAndBroadcastTxSync(msgs []types.Msg) (types.TxResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	txBytes, err := m.TxBuilder.BuildAndSign(m.CLIContext.GetFromName(), m.password, msgs)
	if err != nil {
		return types.TxResponse{}, err
	}

	node, err := m.CLIContext.GetNode()
	if err != nil {
		return types.TxResponse{}, err
	}

	res, err := node.BroadcastTxSync(txBytes)
	if res != nil && res.Code == 0 {
		m.TxBuilder = m.TxBuilder.WithSequence(m.TxBuilder.Sequence() + 1)
	}

	return types.NewResponseFormatBroadcastTx(res), err
}
