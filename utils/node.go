package utils

import (
	"log"

	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ironman0x7b2/sentinel-sdk/x/vpn"
	vpnTypes "github.com/ironman0x7b2/sentinel-sdk/x/vpn/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/common"

	"github.com/ironman0x7b2/vpn-node/config"
	"github.com/ironman0x7b2/vpn-node/tx"
	"github.com/ironman0x7b2/vpn-node/types"
)

func ProcessNode(appCfg *config.AppConfig, tx *tx.Tx, vpn types.BaseVPN) (*vpn.Node, error) {
	fromAddress := tx.Manager.CLIContext.FromAddress
	nodeID := appCfg.Node.ID

	if len(nodeID) == 0 {
		log.Println("Got an empty node ID, so registering node")

		pricesPerGB, err := csdkTypes.ParseCoins(appCfg.Node.PricesPerGB)
		if err != nil {
			return nil, err
		}

		internetSpeed, err := CalculateInternetSpeed()
		if err != nil {
			return nil, err
		}

		msg := vpnTypes.NewMsgRegisterNode(fromAddress,
			appCfg.Node.Moniker, pricesPerGB, internetSpeed,
			vpn.EncryptionMethod(), vpn.Type(), types.Version)

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
	if !node.Owner.Equals(fromAddress) {
		return nil, errors.Errorf("Registered node owner address does not match with current account address")
	}

	return node, nil
}
