package tx

import (
	"encoding/base64"
	"encoding/json"

	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
	"github.com/ironman0x7b2/sentinel-sdk/x/vpn/client/common"
	vpnTypes "github.com/ironman0x7b2/sentinel-sdk/x/vpn/types"
	"github.com/pkg/errors"
	tmTypes "github.com/tendermint/tendermint/types"
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

func (t Tx) CompleteAndSubscribeTx(msgs []csdkTypes.Msg) (*tmTypes.EventDataTx, error) {
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

	return common.QuerySession(t.Manager.CLIContext, t.Manager.CLIContext.Codec, id)
}

func (t Tx) SignSessionBandwidth(id sdkTypes.ID, upload, download int64,
	client csdkTypes.AccAddress) (string, error) {

	bandwidth := sdkTypes.NewBandwidthFromInt64(upload, download)
	bandwidthSignData := sdkTypes.NewBandwidthSignData(id, bandwidth, t.Manager.CLIContext.FromAddress, client)

	msg, err := json.Marshal(bandwidthSignData)
	if err != nil {
		return "", err
	}

	sign, _, err := t.Manager.CLIContext.Keybase.Sign(t.Manager.CLIContext.FromName, t.Manager.password, msg)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sign), nil
}
