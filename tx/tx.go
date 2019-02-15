package tx

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

func CompleteAndSubscribeTx(txManager *Manager, txSubscriber *Subscriber,
	msgs []types.Msg) (tmTypes.EventDataTx, error) {

	res, err := txManager.CompleteAndBroadcastTxSync(msgs)
	if err != nil {
		return tmTypes.EventDataTx{}, err
	}

	c := make(chan tmTypes.EventDataTx)
	defer close(c)

	if err := txSubscriber.WriteTxQuery(res.TxHash, c); err != nil {
		return tmTypes.EventDataTx{}, err
	}

	data := <-c
	if !data.Result.IsOK() {
		return tmTypes.EventDataTx{}, fmt.Errorf(data.Result.String())
	}
	return data, nil
}
