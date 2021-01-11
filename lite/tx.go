package lite

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/sentinel-official/dvpn-node/types"
)

func (c *Client) SignAndBroadcastTxCommit(messages ...sdk.Msg) (*sdk.TxResponse, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	account, err := c.QueryAccount(c.ctx.FromAddress)
	if err != nil {
		return nil, err
	}

	txb := c.txb.WithAccountNumber(account.GetAccountNumber()).
		WithSequence(account.GetSequence())

	if txb.GasAdjustment() > 0 {
		txb, err = utils.EnrichWithGas(txb, c.ctx, messages)
		if err != nil {
			return nil, err
		}
	}

	bytes, err := txb.BuildAndSign(c.ctx.From, types.DefaultPassword, messages)
	if err != nil {
		return nil, err
	}

	node, err := c.ctx.GetNode()
	if err != nil {
		return nil, err
	}

	result, err := node.BroadcastTxCommit(bytes)
	if err != nil {
		return nil, err
	}
	if !result.CheckTx.IsOK() {
		return nil, fmt.Errorf(result.CheckTx.Log)
	}
	if !result.DeliverTx.IsOK() {
		return nil, fmt.Errorf(result.DeliverTx.Log)
	}

	response := sdk.NewResponseFormatBroadcastTxCommit(result)
	return &response, nil
}
