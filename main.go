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
	_cfg := csdk.GetConfig()
	_cfg.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	_cfg.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	_cfg.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	_cfg.Seal()

	cfg := config.NewAppConfig()
	if err := cfg.LoadFromPath(types.DefaultAppConfigFilePath); err != nil {
		panic(err)
	}
	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	log.Printf("Initializing the keybase from directory `%s`", types.DefaultConfigDir)
	kb, err := keys.NewKeyBaseFromDir(types.DefaultConfigDir)
	if err != nil {
		panic(err)
	}

	keyInfo, err := utils.ProcessAccount(kb, cfg.Account.Name)
	if err != nil {
		panic(err)
	}

	cfg.Account.Name = keyInfo.GetName()
	if err = cfg.SaveToPath(types.DefaultAppConfigFilePath); err != nil {
		panic(err)
	}

	password, err := utils.ProcessAccountPassword(kb, cfg.Account.Name)
	if err != nil {
		panic(err)
	}

	ip, err := utils.PublicIP()
	if err != nil {
		panic(err)
	}

	vpn, err := utils.ProcessVPN(cfg.VPNType, ip)
	if err != nil {
		panic(err)
	}

	tx, err := _tx.NewTxWithConfig(cfg.ChainID, cfg.RPCAddress, password, keyInfo, kb)
	if err != nil {
		panic(err)
	}

	nodeInfo, err := utils.ProcessNode(cfg.Node.ID, cfg.Node.Moniker, cfg.Node.PricesPerGB, tx, vpn)
	if err != nil {
		panic(err)
	}

	cfg.Node.ID = nodeInfo.ID.String()
	if err = cfg.SaveToPath(types.DefaultAppConfigFilePath); err != nil {
		panic(err)
	}

	db, err := _db.NewDatabase(types.DefaultDatabaseFilePath)
	if err != nil {
		panic(err)
	}

	if err = utils.NewTLSKey(ip); err != nil {
		panic(err)
	}

	node := _node.NewNode(nodeInfo.ID, keyInfo.GetAddress(), keyInfo.GetPubKey(), tx, db, vpn)
	if err = node.Start(cfg.APIPort); err != nil {
		panic(err)
	}
}
