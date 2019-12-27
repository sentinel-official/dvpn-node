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

func ProcessNode(kb keys.Keybase, cfg *config.AppConfig, tx *_tx.Tx, _vpn types.BaseVPN, password string) (*vpn.Node, error) {
	
	from := tx.Manager.CLI.FromAddress
	if cfg.Node.ID == "" {
		log.Println("Got an empty node ID, so registering the node")
		
		pricesPerGB, err := sdk.ParseCoins(cfg.Node.PricesPerGB)
		if err != nil {
			return nil, err
		}
		
		speed, err := InternetSpeed()
		if err != nil {
			return nil, err
		}
		
		msg := vpn.NewMsgRegisterNode(from, _vpn.Type(), types.Version,
			cfg.Node.Moniker, pricesPerGB, speed, _vpn.Encryption())
		
		fmt.Println(msg)
		
		data, err := tx.CompleteAndSubscribeTx(msg)
		if err != nil {
			return nil, err
		}
		
		events := sdk.StringifyEvents(data.Result.Events)
		cfg.Node.ID = events[1].Attributes[1].Value
		
		log.Printf("Node registered at height `%d`, tx hash `%s`, node ID `%s`",
			data.Height, common.HexBytes(data.Tx.Hash()).String(), cfg.Node.ID)
	}
	
	node, err := tx.QueryNode(cfg.Node.ID)
	if err != nil {
		
		return nil, err
	}
	if !node.Owner.Equals(from) {
		return nil, errors.Errorf("Registered node owner address does not match with current account address")
	}
	
	resolverId, err := hub.NewResolverIDFromString(cfg.Resolver.ID)
	if err != nil {
		panic(err)
	}
	
	resolvers, err := tx.QueryNodesOfResolver(cfg.Resolver.ID)
	if err != nil {
		return nil, err
	}
	
	fmt.Println("nodes at resolver", resolvers)
	isMatch := false
	for _, _resolver := range resolvers {
		if cfg.Node.ID == _resolver.String() {
			isMatch = true
		}
	}
	
	fmt.Println("is match", isMatch)
	if !isMatch {
		log.Println("Node does not find at resolver, so registering the node on resolver")
		
		_msg := vpn.NewMsgRegisterVPNOnResolver(from, node.ID, resolverId)
		
		data, err := tx.CompleteAndSubscribeTx(_msg)
		if err != nil {
			return nil, err
		}
		
		log.Printf("Node registered on resolver at height `%d`, tx hash `%s` resolver-id `%s`",
			data.Height, common.HexBytes(data.Tx.Hash()).String(), resolverId)
	}
	
	ip, err := PublicIP()
	if err != nil {
		return nil, err
	}
	
	sigBytes, pubKey, err := kb.Sign(cfg.Account.Name, password, []byte(nil))
	
	stdSignature := auth.StdSignature{
		PubKey:    pubKey,
		Signature: sigBytes,
	}
	
	_bytes, err := tx.Manager.CLI.Codec.MarshalJSON(stdSignature)
	if err != nil {
		return nil, err
	}
	
	url := "http://" + cfg.Resolver.IP + "/node/register"
	strPort := strconv.FormatUint(uint64(cfg.APIPort), 10)
	message := map[string]interface{}{
		"id":        cfg.Node.ID,
		"ip":        ip,
		"port":      strPort,
		"signature": string(_bytes),
	}
	
	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}
	
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}
	b, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("respppppppppp", string(b))
	if resp.StatusCode != 200 {
		log.Fatalln("Error while register on the resolver")
	}
	
	log.Fatalln("Register node on the resolver completed locally")
	
	return node, nil
}
