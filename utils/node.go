package utils

import (
	"log"

	csdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/common"

	"github.com/ironman0x7b2/sentinel-sdk/x/vpn"

	"github.com/ironman0x7b2/vpn-node/config"
	_tx "github.com/ironman0x7b2/vpn-node/tx"
	"github.com/ironman0x7b2/vpn-node/types"
)

func ProcessNode(appCfg *config.AppConfig, tx *_tx.Tx, _vpn types.BaseVPN) (*vpn.Node, error) {
	nodeID := appCfg.Node.ID

	if nodeID == "" {
		log.Println("Got an empty node ID, so registering the node")

		pricesPerGB, err := csdk.ParseCoins(appCfg.Node.PricesPerGB)
		if err != nil {
			return nil, err
		}

		internetSpeed, err := CalculateInternetSpeed()
		if err != nil {
			return nil, err
		}

		msg := vpn.NewMsgRegisterNode(tx.FromAddress(), _vpn.Type(), types.Version,
			appCfg.Node.Moniker, pricesPerGB, internetSpeed, _vpn.Encryption())

		data, err := tx.CompleteAndSubscribeTx(msg)
		if err != nil {
			return nil, err
		}

		nodeID = string(data.Result.Tags[1].Value)

		log.Printf("Node registered at height `%d`, tx hash `%s`, node ID `%s`",
			data.Height, common.HexBytes(data.Tx.Hash()).String(), nodeID)
	}

	node, err := tx.QueryNode(nodeID)
	if err != nil {
		return nil, err
	}
	if !node.Owner.Equals(tx.FromAddress()) {
		return nil, errors.Errorf("Registered node owner address does not match with current account address")
	}

	return node, nil
}
