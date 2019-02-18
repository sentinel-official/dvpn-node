package tx

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ironman0x7b2/sentinel-sdk/types"
	"github.com/ironman0x7b2/sentinel-sdk/x/vpn/client/common"
	vpnTypes "github.com/ironman0x7b2/sentinel-sdk/x/vpn/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

type Tx struct {
	manager    *Manager
	subscriber *Subscriber
}

func NewTx(manager *Manager, subscriber *Subscriber) *Tx {
	return &Tx{
		manager:    manager,
		subscriber: subscriber,
	}
}

func (t Tx) CompleteAndSubscribeTx(msgs []csdkTypes.Msg) (*tmTypes.EventDataTx, error) {
	res, err := t.manager.CompleteAndBroadcastTxSync(msgs)
	if err != nil {
		return nil, err
	}

	c := make(chan tmTypes.EventDataTx)
	defer close(c)

	if err := t.subscriber.WriteTxQuery(res.TxHash, c); err != nil {
		return nil, err
	}

	data := <-c
	if !data.Result.IsOK() {
		return nil, fmt.Errorf(data.Result.String())
	}

	return &data, nil
}

func (t Tx) QuerySessionFromTxHash(txHash string) (*vpnTypes.SessionDetails, error) {
	res, err := t.manager.QueryTx(txHash)
	if err != nil {
		return nil, err
	}

	if !res.TxResult.IsOK() {
		return nil, fmt.Errorf(res.TxResult.Log)
	}

	id := string(res.TxResult.Tags[2].Value)

	return common.QuerySession(t.manager.CLIContext, t.manager.CLIContext.Codec, id)
}

func (t Tx) SignSessionBandwidth(id vpnTypes.SessionID, upload, download int64,
	client csdkTypes.AccAddress) (string, error) {

	bandwidth := types.NewBandwidthFromInt64(upload, download)
	bandwidthSign := vpnTypes.NewBandwidthSign(id, bandwidth, t.manager.CLIContext.FromAddress, client)

	msg, err := json.Marshal(bandwidthSign)
	if err != nil {
		return "", err
	}

	sign, _, err := t.manager.CLIContext.Keybase.Sign(t.manager.CLIContext.FromName, t.manager.password, msg)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sign), nil
}
