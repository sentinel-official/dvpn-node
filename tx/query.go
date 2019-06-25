package tx

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/pkg/errors"

	"github.com/sentinel-official/sentinel-hub/x/vpn"
	"github.com/sentinel-official/sentinel-hub/x/vpn/client/common"
)

func (t Tx) QueryAccount(_address string) (auth.Account, error) {
	address, err := sdk.AccAddressFromBech32(_address)
	if err != nil {
		return nil, err
	}

	account, err := t.Manager.CLIContext.GetAccount(address)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, errors.Errorf("no account found")
	}

	return account, nil
}

func (t Tx) QueryNode(id string) (*vpn.Node, error) {
	return common.QueryNode(t.Manager.CLIContext, t.Manager.CLIContext.Codec, id)
}

func (t Tx) QuerySubscription(id string) (*vpn.Subscription, error) {
	return common.QuerySubscription(t.Manager.CLIContext, t.Manager.CLIContext.Codec, id)
}

func (t Tx) QuerySubscriptionByTxHash(hash string) (*vpn.Subscription, error) {
	res, err := t.Manager.QueryTx(hash)
	if err != nil {
		return nil, err
	}
	if !res.TxResult.IsOK() {
		return nil, errors.Errorf(res.TxResult.String())
	}

	var stdTx auth.StdTx
	if err := t.Manager.CLIContext.Codec.UnmarshalBinaryLengthPrefixed(res.Tx, &stdTx); err != nil {
		return nil, err
	}

	if len(stdTx.Msgs) != 1 || stdTx.Msgs[0].Type() != "MsgStartSubscription" {
		return nil, errors.Errorf("Invalid subscription transaction")
	}

	id := string(res.TxResult.Tags[2].Value)
	return common.QuerySubscription(t.Manager.CLIContext, t.Manager.CLIContext.Codec, id)
}

func (t Tx) QuerySessionsCountOfSubscription(id string) (uint64, error) {
	return common.QuerySessionsCountOfSubscription(t.Manager.CLIContext, t.Manager.CLIContext.Codec, id)
}

func (t Tx) QuerySessionOfSubscription(id string, index uint64) (*vpn.Session, error) {
	return common.QuerySessionOfSubscription(t.Manager.CLIContext, t.Manager.CLIContext.Codec, id, index)
}
