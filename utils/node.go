package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/pkg/errors"
	hub "github.com/sentinel-official/hub/types"
	vpn "github.com/sentinel-official/hub/x/vpn/types"
	"github.com/tendermint/tendermint/libs/common"
	
	"github.com/sentinel-official/dvpn-node/config"
	_tx "github.com/sentinel-official/dvpn-node/tx"
	"github.com/sentinel-official/dvpn-node/types"
)

func ProcessNode(cfg *config.AppConfig, tx *_tx.Tx, _vpn types.BaseVPN) (*config.AppConfig, *vpn.Node, error) {
	from := tx.Manager.CLI.FromAddress
	if cfg.Node.ID == "" {
		log.Println("Got an empty node ID, so registering the node")
		
		pricesPerGB, err := sdk.ParseCoins(cfg.Node.PricesPerGB)
		if err != nil {
			return cfg, nil, err
		}
		
		speed, err := InternetSpeed()
		if err != nil {
			return cfg, nil, err
		}
		
		msg := vpn.NewMsgRegisterNode(from, _vpn.Type(), types.Version,
			cfg.Node.Moniker, pricesPerGB, speed, _vpn.Encryption())
		
		fmt.Println(msg)
		
		data, err := tx.CompleteAndSubscribeTx(msg)
		if err != nil {
			return cfg, nil, err
		}
		
		events := sdk.StringifyEvents(data.Result.Events)
		cfg.Node.ID = events[1].Attributes[1].Value
		
		log.Printf("Node registered at height `%d`, tx hash `%s`, node ID `%s`",
			data.Height, common.HexBytes(data.Tx.Hash()).String(), cfg.Node.ID)
	}
	
	node, err := tx.QueryNode(cfg.Node.ID)
	if err != nil {
		return cfg, nil, err
	}
	if !node.Owner.Equals(from) {
		return cfg, nil, errors.Errorf("Registered node owner address does not match with current account address")
	}
	
	return cfg, node, nil
}

func ProcessResolver(kb keys.Keybase, cfg *config.AppConfig, tx *_tx.Tx, nodeID hub.NodeID, password string) error {
	from := tx.Manager.CLI.FromAddress
	resolverId, err := hub.NewResolverIDFromString(cfg.Resolver.ID)
	if err != nil {
		panic(err)
	}
	
	resolvers, err := tx.QueryNodesOfResolver(cfg.Resolver.ID)
	if err != nil {
		return err
	}
	
	isRegister := false
	for _, _resolver := range resolvers {
		if cfg.Node.ID == _resolver.String() {
			isRegister = true
		}
	}
	
	if !isRegister {
		log.Println("Node does not register at resolver, so registering the node at resolver")
		
		_msg := vpn.NewMsgRegisterVPNOnResolver(from, nodeID, resolverId)
		
		data, err := tx.CompleteAndSubscribeTx(_msg)
		if err != nil {
			return err
		}
		
		log.Printf("Node registered on resolver at height `%d`, tx hash `%s` resolver-id `%s`",
			data.Height, common.HexBytes(data.Tx.Hash()).String(), resolverId)
	}
	
	url := "http://" + cfg.Resolver.IP + "/nodes/" + nodeID.String()
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	
	var resp types.Response
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return err
	}
	
	bz, err := json.Marshal(resp.Result)
	if err != nil {
		return err
	}
	
	var node types.Node
	err = json.Unmarshal(bz, &node)
	if err != nil {
		return err
	}
	
	if node.ID == "" {
		log.Println("Node does not register at local resolver, so registering the node at local resolver")
		
		ip, err := PublicIP()
		if err != nil {
			return err
		}
		
		sigBytes, pubKey, err := kb.Sign(cfg.Account.Name, password, []byte(nil))
		
		stdSignature := auth.StdSignature{
			PubKey:    pubKey,
			Signature: sigBytes,
		}
		
		_bytes, err := tx.Manager.CLI.Codec.MarshalJSON(stdSignature)
		if err != nil {
			return err
		}
		
		url := "http://" + cfg.Resolver.IP + "/node/register"
		strPort := strconv.FormatUint(uint64(cfg.APIPort), 10)
		
		message := map[string]interface{}{
			"id":        cfg.Node.ID,
			"ip":        ip,
			"port":      strPort,
			"signature": string(_bytes),
		}
		
		bz, err := json.Marshal(message)
		if err != nil {
			log.Fatalln(err)
		}
		
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(bz))
		if err != nil {
			log.Fatalln(err)
		}
		
		if resp.StatusCode != 200 {
			log.Fatalln("Error while register on the local resolver")
		}
		
		log.Fatalln("Register node on the local resolver completed")
	}
	
	return nil
}
