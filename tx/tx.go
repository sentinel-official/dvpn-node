package tx

import (
	"encoding/base64"
	"encoding/json"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client/context"
	clientUtils "github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	clientTxBuilder "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/ironman0x7b2/sentinel-sdk/apps/vpn"
	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
	"github.com/ironman0x7b2/sentinel-sdk/x/vpn/client/common"
	vpnTypes "github.com/ironman0x7b2/sentinel-sdk/x/vpn/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/lite/proxy"
	"github.com/tendermint/tendermint/rpc/client"
	tmTypes "github.com/tendermint/tendermint/types"

	"github.com/ironman0x7b2/vpn-node/config"
	"github.com/ironman0x7b2/vpn-node/types"
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

	verifier, err := proxy.NewVerifier(appCfg.ChainID, filepath.Join(types.DefaultConfigDir, ".vpn_lite"),
		client.NewHTTP(appCfg.LiteClientURI, "/websocket"), log.NewNopLogger(), 10)
	if err != nil {
		return nil, err
	}

	cliContext := context.NewCLIContext().
		WithCodec(cdc).
		WithAccountDecoder(cdc).
		WithNodeURI(appCfg.LiteClientURI).
		WithVerifier(verifier).
		WithFrom(ownerInfo.GetName()).
		WithFromName(ownerInfo.GetName()).
		WithFromAddress(ownerInfo.GetAddress())

	account, err := cliContext.GetAccount(ownerInfo.GetAddress().Bytes())
	if err != nil {
		return nil, err
	}

	txBuilder := clientTxBuilder.NewTxBuilderFromCLI().
		WithKeybase(keybase).
		WithAccountNumber(account.GetAccountNumber()).
		WithSequence(account.GetSequence()).
		WithChainID(appCfg.ChainID).
		WithGas(1000000000).
		WithTxEncoder(clientUtils.GetTxEncoder(cdc))

	manager := NewManager(cliContext, txBuilder, appCfg.Owner.Password)
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

func (t Tx) QueryNodeDetails(id string) (*vpnTypes.NodeDetails, error) {
	return common.QueryNode(t.Manager.CLIContext, t.Manager.CLIContext.Codec, sdkTypes.NewID(id))
}

func (t Tx) OwnerAddress() csdkTypes.AccAddress {
	return t.Manager.CLIContext.FromAddress
}
