package main

import (
	"log"

	"github.com/cosmos/cosmos-sdk/client/keys"
	csdk "github.com/cosmos/cosmos-sdk/types"

	sdk "github.com/ironman0x7b2/sentinel-sdk/types"

	"github.com/ironman0x7b2/vpn-node/config"
	_db "github.com/ironman0x7b2/vpn-node/db"
	_node "github.com/ironman0x7b2/vpn-node/node"
	_tx "github.com/ironman0x7b2/vpn-node/tx"
	"github.com/ironman0x7b2/vpn-node/types"
	"github.com/ironman0x7b2/vpn-node/utils"
)

// nolint:gocyclo
func main() {
	cfg := csdk.GetConfig()
	cfg.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	cfg.Seal()

	appCfg := config.NewAppConfig()
	if err := appCfg.LoadFromPath(types.DefaultAppConfigFilePath); err != nil {
		panic(err)
	}

	log.Printf("Initializing the keybase from directory `%s`", types.DefaultConfigDir)
	kb, err := keys.NewKeyBaseFromDir(types.DefaultConfigDir)
	if err != nil {
		panic(err)
	}

	info, err := utils.ProcessAccount(kb, appCfg.Account.Name)
	if err != nil {
		panic(err)
	}

	appCfg.Account.Name = info.GetName()
	if err = appCfg.SaveToPath(types.DefaultAppConfigFilePath); err != nil {
		panic(err)
	}

	appCfg.Account.Password, err = utils.ProcessAccountPassword(kb, appCfg.Account.Name)
	if err != nil {
		panic(err)
	}

	vpn, err := utils.ProcessVPN(appCfg.VPNType)
	if err != nil {
		panic(err)
	}

	tx, err := _tx.NewTxFromConfig(appCfg, info, kb)
	if err != nil {
		panic(err)
	}

	node, err := utils.ProcessNode(appCfg, tx, vpn)
	if err != nil {
		panic(err)
	}

	appCfg.Node.ID = node.ID.String()
	if err = appCfg.SaveToPath(types.DefaultAppConfigFilePath); err != nil {
		panic(err)
	}

	db, err := _db.NewDatabase("sentinel.db")
	if err != nil {
		panic(err)
	}

	_node.NewNode(node.ID, info.GetAddress(), info.GetPubKey(),
		tx, vpn, db).Start(appCfg.APIPort)
}
