package main

import (
	"github.com/cosmos/cosmos-sdk/client/keys"

	"github.com/ironman0x7b2/vpn-node/config"
	"github.com/ironman0x7b2/vpn-node/node"
	"github.com/ironman0x7b2/vpn-node/tx"
	"github.com/ironman0x7b2/vpn-node/types"
	"github.com/ironman0x7b2/vpn-node/utils"
)

func main() {
	appCfg := config.NewAppConfig()
	if err := appCfg.LoadFromPath(""); err != nil {
		panic(err)
	}

	defer func() {
		if err := appCfg.SaveToPath(""); err != nil {
			panic(err)
		}
	}()

	keybase, err := keys.NewKeyBaseFromDir(types.DefaultConfigDir)
	if err != nil {
		panic(err)
	}

	ownerInfo, err := utils.ProcessOwnerAccount(keybase, appCfg.Owner.Name)
	if err != nil {
		panic(err)
	}

	appCfg.Owner.Name = ownerInfo.GetName()
	appCfg.Owner.Address = ownerInfo.GetAddress().String()

	password, err := utils.ProcessAccountPassword(keybase, appCfg.Owner.Name)
	if err != nil {
		panic(err)
	}

	appCfg.Owner.Password = password

	_tx, err := tx.NewTxFromConfig(appCfg, ownerInfo, keybase)
	if err != nil {
		panic(err)
	}

	vpn, err := utils.ProcessVPN(appCfg.VPNType)
	if err != nil {
		panic(err)
	}

	details, err := utils.ProcessNodeDetails(appCfg, _tx, vpn)
	if err != nil {
		panic(err)
	}

	appCfg.Node.ID = details.ID.String()

	_node := node.NewNode(details, _tx, vpn)
	if err := _node.Start(); err != nil {
		panic(err)
	}
}
