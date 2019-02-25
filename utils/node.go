package utils

import (
	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	vpnTypes "github.com/ironman0x7b2/sentinel-sdk/x/vpn/types"
	"github.com/pkg/errors"

	"github.com/ironman0x7b2/vpn-node/config"
	"github.com/ironman0x7b2/vpn-node/tx"
	"github.com/ironman0x7b2/vpn-node/types"
)

func ProcessNodeDetails(appCfg *config.AppConfig, tx *tx.Tx, vpn types.BaseVPN) (*vpnTypes.NodeDetails, error) {
	nodeID := appCfg.Node.ID

	if len(nodeID) == 0 {
		amountToLock, err := csdkTypes.ParseCoin(appCfg.Node.AmountToLock)
		if err != nil {
			return nil, err
		}

		pricesPerGB, err := csdkTypes.ParseCoins(appCfg.Node.PricesPerGB)
		if err != nil {
			return nil, err
		}

		netSpeed, err := CalculateInternetSpeed()
		if err != nil {
			return nil, err
		}

		apiPort := vpnTypes.NewAPIPort(appCfg.Node.APIPort)

		msg := vpnTypes.NewMsgRegisterNode(tx.OwnerAddress(), amountToLock, pricesPerGB,
			netSpeed.Upload, netSpeed.Download, apiPort, vpn.Encryption(), vpn.Type(), types.Version)
		data, err := tx.CompleteAndSubscribeTx(msg)
		if err != nil {
			return nil, err
		}

		nodeID = string(data.Result.Tags[2].Value)
	}

	details, err := tx.QueryNodeDetails(nodeID)
	if err != nil {
		return nil, err
	}
	if !details.Owner.Equals(tx.OwnerAddress()) {
		return nil, errors.New("Node owner mismatch")
	}

	return details, nil
}
