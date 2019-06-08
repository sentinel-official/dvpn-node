package tx

import (
	"log"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	csdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	tm "github.com/tendermint/tendermint/types"

	"github.com/ironman0x7b2/sentinel-sdk/app/hub"

	"github.com/ironman0x7b2/vpn-node/config"
)

type Tx struct {
	Manager    *Manager
	Subscriber *Subscriber
}

func NewTx(manager *Manager, subscriber *Subscriber) *Tx {
	return &Tx{
		Manager:    manager,
		Subscriber: subscriber,
	}
}

func NewTxFromConfig(appCfg *config.AppConfig, info keys.Info, kb keys.Keybase) (*Tx, error) {
	cdc := hub.MakeCodec()
	tm.RegisterEventDatas(cdc)

	log.Println("Initializing the transaction manager")
	manager, err := NewManagerFromConfig(appCfg, cdc, info, kb)
	if err != nil {
		return nil, err
	}

	log.Println("Initializing the transaction subscriber")
	subscriber, err := NewSubscriber(appCfg.RPCAddress, cdc)
	if err != nil {
		return nil, err
	}

	return NewTx(manager, subscriber), nil
}

func (t Tx) CompleteAndSubscribeTx(messages ...csdk.Msg) (*tm.EventDataTx, error) {
	res, err := t.Manager.CompleteAndBroadcastTxSync(messages)
	if err != nil {
		return nil, err
	}

	event := make(chan tm.EventDataTx)
	defer close(event)

	if err := t.Subscriber.WriteTxQuery(res.TxHash, event); err != nil {
		return nil, err
	}

	data := <-event
	if !data.Result.IsOK() {
		return nil, errors.Errorf(data.Result.String())
	}

	return &data, nil
}
