package core

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/sentinel-official/sentinel-go-sdk/v2/libs/log"
)

// BroadcastTx safely broadcasts a transaction with the provided messages.
// It locks the transaction mutex to ensure only one transaction is broadcast at a time.
func (c *Context) BroadcastTx(ctx context.Context, msgs ...types.Msg) error {
	c.txm.Lock()
	defer c.txm.Unlock()

	// No messages to broadcast, skipping.
	if len(msgs) == 0 {
		return nil
	}

	// Broadcast the transaction and wait for it to be included in a block.
	txResp, txRes, err := c.Client().BroadcastTxCommit(ctx, msgs...)
	if err != nil {
		return fmt.Errorf("broadcasting tx commit: %w", err)
	}

	log.Debug(
		"Transaction broadcasted successfully",
		"code", fmt.Sprintf("%s/%d", txRes.TxResult.Codespace, txRes.TxResult.Code),
		"gas", fmt.Sprintf("%d/%d", txRes.TxResult.GasUsed, txRes.TxResult.GasWanted),
		"hash", txResp.Hash,
		"height", txRes.Height,
		"msgs", len(msgs),
	)

	return nil
}
