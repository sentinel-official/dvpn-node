package tx

import (
	"encoding/base64"
	"log"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ironman0x7b2/sentinel-sdk/apps/vpn"
	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
	"github.com/ironman0x7b2/sentinel-sdk/x/vpn/client/common"
	vpnTypes "github.com/ironman0x7b2/sentinel-sdk/x/vpn/types"
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

func NewTxFromConfig(appCfg *config.AppConfig, ownerInfo keys.Info, keybase keys.Keybase) (*Tx, error) {
	cdc := vpn.MakeCodec()
	tmTypes.RegisterEventDatas(cdc)

	log.Println("Initializing the transaction manager")
	manager, err := NewManagerFromConfig(appCfg, cdc, ownerInfo, keybase)
	if err != nil {
		return nil, err
	}

	log.Println("Initializing the transaction subscriber")
	subscriber, err := NewSubscriber(appCfg.LiteClientURI, cdc)
	if err != nil {
		return nil, err
	}

	return NewTx(manager, subscriber), nil
}

func (t Tx) CompleteAndSubscribeTx(msgs ...csdkTypes.Msg) (*tmTypes.EventDataTx, error) {
	res, err := t.Manager.CompleteAndBroadcastTxSync(msgs)
	if err != nil {
		return nil, err
	}

	c := make(chan tmTypes.EventDataTx)
	defer close(c)

	if err := t.Subscriber.WriteTxQuery(res.TxHash, c); err != nil {
		return nil, err
	}

	data := <-c
	if !data.Result.IsOK() {
		return nil, errors.New(data.Result.String())
	}

	return &data, nil
}

func (t Tx) QuerySessionDetailsFromTxHash(txHash string) (*vpnTypes.SessionDetails, error) {
	res, err := t.Manager.QueryTx(txHash)
	if err != nil {
		return nil, err
	}

	if !res.TxResult.IsOK() {
		return nil, errors.New(res.TxResult.String())
	}

	id := string(res.TxResult.Tags[2].Value)

	log.Printf("Querying the session details with session ID `%s`", id)
	return common.QuerySession(t.Manager.CLIContext, t.Manager.CLIContext.Codec, id)
}

func (t Tx) SignSessionBandwidth(id sdkTypes.ID, bandwidth sdkTypes.Bandwidth,
	client csdkTypes.AccAddress) (string, error) {

	msg := sdkTypes.NewBandwidthSignData(id, bandwidth, t.Manager.CLIContext.FromAddress, client).GetBytes()

	log.Printf("Signing the bandwidth details of session with ID `%s`", id.String())
	sign, _, err := t.Manager.CLIContext.Keybase.Sign(t.Manager.CLIContext.FromName, t.Manager.password, msg)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sign), nil
}

func (t Tx) QueryNodeDetails(id string) (*vpnTypes.NodeDetails, error) {
	nodeID := sdkTypes.NewID(id)

	log.Printf("Querying the node details with node ID `%s`", id)
	return common.QueryNode(t.Manager.CLIContext, t.Manager.CLIContext.Codec, nodeID)
}
