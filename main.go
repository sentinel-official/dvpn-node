package main

import (
	"log"

	"github.com/cosmos/cosmos-sdk/client/keys"

	"github.com/ironman0x7b2/vpn-node/config"
	"github.com/ironman0x7b2/vpn-node/node"
	"github.com/ironman0x7b2/vpn-node/tx"
	"github.com/ironman0x7b2/vpn-node/types"
	"github.com/ironman0x7b2/vpn-node/utils"
)

func main() {
	appCfg := config.NewAppConfig()
	if err := appCfg.LoadFromPath(types.DefaultAppConfigFilePath); err != nil {
		panic(err)
	}

	log.Printf("Initializing the keybase from directory `%s`", types.DefaultConfigDir)
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

	if err := appCfg.SaveToPath(types.DefaultAppConfigFilePath); err != nil {
		panic(err)
	}

	password, err := utils.ProcessAccountPassword(keybase, appCfg.Owner.Name)
	if err != nil {
		panic(err)
	}

	appCfg.Owner.Password = password

	vpn, err := utils.ProcessVPN(appCfg.VPNType)
	if err != nil {
		panic(err)
	}

	_tx, err := tx.NewTxFromConfig(appCfg, ownerInfo, keybase)
	if err != nil {
		panic(err)
	}

	details, err := utils.ProcessNodeDetails(appCfg, _tx, vpn)
	if err != nil {
		panic(err)
	}

	appCfg.Node.ID = details.ID.String()

	if err := appCfg.SaveToPath(types.DefaultAppConfigFilePath); err != nil {
		panic(err)
	}

	log.Printf("Initializing the node")
	_node := node.NewNode(details, _tx, vpn)

	if err := _node.Start(); err != nil {
		panic(err)
	}
}
