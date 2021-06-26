package lite

import (
	"github.com/avast/retry-go"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/pkg/errors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

func (c *Client) prepareTxFactory(messages ...sdk.Msg) (txf tx.Factory, err error) {
	defer func() {
		c.Log().Error("Failed to prepare the transaction", "error", err)
	}()

	account, err := c.QueryAccount(c.FromAddress())
	if err != nil {
		return txf, err
	}

	txf = c.txf.
		WithAccountNumber(account.GetAccountNumber()).
		WithSequence(account.GetSequence())

	if c.SimulateAndExecute() {
		_, adjusted, err := tx.CalculateGas(c.ctx.QueryWithData, txf, messages...)
		if err != nil {
			return txf, err
		}

		txf = txf.WithGas(adjusted)
	}

	return txf, nil
}

func (c *Client) broadcastTx(txBytes []byte) (res *sdk.TxResponse, err error) {
	defer func() {
		c.Log().Error("Failed to broadcast the transaction", "error", err)
	}()

	res, err = c.ctx.BroadcastTx(txBytes)
	if err != nil {
		return nil, err
	}

	switch res.Code {
	case abcitypes.CodeTypeOK:
		return res, nil
	case sdkerrors.ErrTxInMempoolCache.ABCICode():
		return res, nil
	default:
		return nil, errors.New(res.RawLog)
	}
}

func (c *Client) Tx(messages ...sdk.Msg) (res *sdk.TxResponse, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var (
		txf tx.Factory
	)

	c.Log().Info("Preparing the transaction", "messages", len(messages))
	if err := retry.Do(func() error {
		txf, err = c.prepareTxFactory(messages...)
		if err != nil {
			return err
		}

		c.Log().Info("Transaction info", "gas", txf.Gas(), "sequence", txf.Sequence())
		return nil
	}, retry.Attempts(5)); err != nil {
		return nil, err
	}

	txb, err := tx.BuildUnsignedTx(txf, messages...)
	if err != nil {
		return nil, err
	}

	if err := tx.Sign(txf, c.From(), txb, true); err != nil {
		return nil, err
	}

	txBytes, err := c.TxConfig().TxEncoder()(txb.GetTx())
	if err != nil {
		return nil, err
	}

	c.Log().Info("Broadcasting the transaction", "size", len(txBytes))
	if err := retry.Do(func() error {
		res, err = c.broadcastTx(txBytes)
		if err != nil {
			return err
		}

		c.Log().Info("Transaction result", "code", res.Code,
			"codespace", res.Codespace, "height", res.Height, "tx_hash", res.TxHash)
		return nil
	}, retry.Attempts(5)); err != nil {
		return nil, err
	}

	return res, nil
}
