package tx

import (
	"encoding/base64"
	"log"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ironman0x7b2/sentinel-sdk/app/hub"
	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
	"github.com/ironman0x7b2/sentinel-sdk/x/vpn"
	"github.com/ironman0x7b2/sentinel-sdk/x/vpn/client/common"
	"github.com/pkg/errors"
	tmTypes "github.com/tendermint/tendermint/types"

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

func NewTxFromConfig(appCfg *config.AppConfig, ownerInfo keys.Info, kb keys.Keybase) (*Tx, error) {
	cdc := hub.MakeCodec()
	tmTypes.RegisterEventDatas(cdc)

	log.Println("Initializing the transaction manager")
	manager, err := NewManagerFromConfig(appCfg, cdc, ownerInfo, kb)
	if err != nil {
		return nil, err
	}

	log.Println("Initializing the transaction subscriber")
	subscriber, err := NewSubscriber(appCfg.RPCServerAddress, cdc)
	if err != nil {
		return nil, err
	}

	return NewTx(manager, subscriber), nil
}

func (t Tx) CompleteAndSubscribeTx(messages ...csdkTypes.Msg) (*tmTypes.EventDataTx, error) {
	res, err := t.Manager.CompleteAndBroadcastTxSync(messages)
	if err != nil {
		return nil, err
	}

	event := make(chan tmTypes.EventDataTx)
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

func (t Tx) QuerySessionFromTxHash(hash string) (*vpn.Session, error) {
	res, err := t.Manager.QueryTx(hash)
	if err != nil {
		return nil, err
	}

	if !res.TxResult.IsOK() {
		return nil, errors.Errorf(res.TxResult.String())
	}

	id := string(res.TxResult.Tags[2].Value)

	log.Printf("Querying the session with session ID `%s`", id)
	return common.QuerySession(t.Manager.CLIContext, t.Manager.CLIContext.Codec, id)
}

func (t Tx) SignSessionBandwidth(id sdkTypes.ID, bandwidth sdkTypes.Bandwidth,
	client csdkTypes.AccAddress) (string, error) {

	data := sdkTypes.NewBandwidthSign(id, bandwidth, t.Manager.CLIContext.FromAddress, client).GetBytes()
	sign, _, err := t.Manager.CLIContext.Keybase.Sign(t.Manager.CLIContext.FromName, t.Manager.password, data)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sign), nil
}

func (t Tx) QueryNode(id string) (*vpn.Node, error) {
	nodeID := sdkTypes.NewID(id)

	log.Printf("Querying the node with node ID `%s`", id)
	return common.QueryNode(t.Manager.CLIContext, t.Manager.CLIContext.Codec, nodeID)
}
