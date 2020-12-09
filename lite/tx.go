package lite

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/sentinel-official/dvpn-node/types"
)

func (c *Client) SignAndBroadcastTxCommit(messages ...sdk.Msg) (sdk.TxResponse, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	account, err := c.QueryAccount(c.ctx.FromAddress)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	txb := c.txb.WithAccountNumber(account.GetAccountNumber()).
		WithSequence(account.GetSequence())

	if txb.GasAdjustment() > 0 {
		txb, err = utils.EnrichWithGas(txb, c.ctx, messages)
		if err != nil {
			return sdk.TxResponse{}, err
		}
	}

	bytes, err := txb.BuildAndSign(c.ctx.From, types.DefaultPassword, messages)
	if err != nil {
		return sdk.TxResponse{}, err
	}

	node, err := c.ctx.GetNode()
	if err != nil {
		return sdk.TxResponse{}, err
	}

	result, err := node.BroadcastTxCommit(bytes)
	if err != nil {
		return sdk.TxResponse{}, err
	}
	if !result.CheckTx.IsOK() {
		return sdk.TxResponse{}, fmt.Errorf(result.CheckTx.Log)
	}
	if !result.DeliverTx.IsOK() {
		return sdk.TxResponse{}, fmt.Errorf(result.DeliverTx.Log)
	}

	return sdk.NewResponseFormatBroadcastTxCommit(result), nil
}
