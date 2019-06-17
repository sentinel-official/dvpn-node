package tx

import (
	"log"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	csdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	tm "github.com/tendermint/tendermint/types"

	"github.com/ironman0x7b2/sentinel-sdk/app/hub"
	sdk "github.com/ironman0x7b2/sentinel-sdk/types"
	"github.com/ironman0x7b2/sentinel-sdk/x/vpn"
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

func NewTxWithConfig(chainID, rpcAddress, password string, keyInfo keys.Info, kb keys.Keybase) (*Tx, error) {
	cdc := hub.MakeCodec()
	tm.RegisterEventDatas(cdc)

	log.Println("Initializing the transaction manager")
	manager, err := NewManagerWithConfig(chainID, rpcAddress, password, cdc, keyInfo, kb)
	if err != nil {
		return nil, err
	}

	log.Println("Initializing the transaction subscriber")
	subscriber, err := NewSubscriber(rpcAddress, cdc)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := subscriber.ReadTxQuery(); err != nil {
			panic(err)
		}
	}()

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

func (t Tx) SignSessionBandwidth(id sdk.ID, index uint64, bandwidth sdk.Bandwidth) ([]byte, error) {
	signature, _, err := t.Manager.CLIContext.Keybase.Sign(t.Manager.CLIContext.FromName,
		t.Manager.password, vpn.NewBandwidthSignatureData(id, index, bandwidth).Bytes())
	if err != nil {
		return nil, err
	}

	return signature, nil
}
