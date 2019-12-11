package utils

import (
	"log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	vpn "github.com/sentinel-official/hub/x/vpn/types"
	"github.com/tendermint/tendermint/libs/common"

	_tx "github.com/sentinel-official/dvpn-node/tx"
	"github.com/sentinel-official/dvpn-node/types"
)

func ProcessNode(id, moniker, _pricesPerGB string, tx *_tx.Tx, _vpn types.BaseVPN, resolver sdk.AccAddress) (*vpn.Node, error) {
	from := tx.Manager.CLI.FromAddress

	if id == "" {
		log.Println("Got an empty node ID, so registering the node")

		pricesPerGB, err := sdk.ParseCoins(_pricesPerGB)
		if err != nil {
			return nil, err
		}

		speed, err := InternetSpeed()
		if err != nil {
			return nil, err
		}

		msg := vpn.NewMsgRegisterNode(from, _vpn.Type(), types.Version,
			moniker, pricesPerGB, speed, _vpn.Encryption())

		data, err := tx.CompleteAndSubscribeTx(msg)
		if err != nil {
			return nil, err
		}

		id = data.Result.Events[1].String()

		log.Printf("Node registered at height `%d`, tx hash `%s`, node ID `%s`",
			data.Height, common.HexBytes(data.Tx.Hash()).String(), id)
	}

	node, err := tx.QueryNode(id)
	if err != nil {
		return nil, err
	}
	if !node.Owner.Equals(from) {
		return nil, errors.Errorf("Registered node owner address does not match with current account address")
	}

	_msg := vpn.NewMsgAddVPNOnResolver(from, node.ID, resolver)

	data, err := tx.CompleteAndSubscribeTx(_msg)
	if err != nil {
		return nil, err
	}

	log.Printf("Node registered on resolver at height `%d`, tx hash `%s`",
		data.Height, common.HexBytes(data.Tx.Hash()).String())

	resolvers, err := tx.QueryResolver(id)
	if err != nil {
		return nil, err
	}

	isMatch := false
	for _, resolver := range resolvers {
		if node.Owner.Equals(resolver) {
			isMatch = true
		}
	}

	if !isMatch {
		return nil, errors.Errorf("Registered node owner address does not match with resolver address")
	}

	return node, nil
}
