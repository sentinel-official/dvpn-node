package utils

import (
	"log"

	csdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/common"

	"github.com/ironman0x7b2/sentinel-sdk/x/vpn"

	_tx "github.com/ironman0x7b2/vpn-node/tx"
	"github.com/ironman0x7b2/vpn-node/types"
)

func ProcessNode(id, moniker, _pricesPerGB string, tx *_tx.Tx, _vpn types.BaseVPN) (*vpn.Node, error) {
	from := tx.Manager.CLIContext.FromAddress

	if id == "" {
		log.Println("Got an empty node ID, so registering the node")

		pricesPerGB, err := csdk.ParseCoins(_pricesPerGB)
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

		id = string(data.Result.Tags[1].Value)

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

	return node, nil
}
