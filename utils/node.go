package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	vpn "github.com/sentinel-official/hub/x/vpn/types"
	"github.com/tendermint/tendermint/libs/common"

	_tx "github.com/sentinel-official/dvpn-node/tx"
	"github.com/sentinel-official/dvpn-node/types"
)

func ProcessNode(id, moniker, _pricesPerGB string, tx *_tx.Tx, _vpn types.BaseVPN,
	resolver sdk.AccAddress, resolverIP string, port uint16) (*vpn.Node, error) {

	from := tx.Manager.CLI.FromAddress
	if id == "" {
		log.Println("Got an empty node ID, so registering the node")

		pricesPerGB, err := sdk.ParseCoins(_pricesPerGB)
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

		events := sdk.StringifyEvents(data.Result.Events)
		id = events[1].Attributes[1].Value

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

	_msg := vpn.NewMsgAddVPNOnResolver(from, node.ID, resolver)

	data, err := tx.CompleteAndSubscribeTx(_msg)
	if err != nil {
		return nil, err
	}

	log.Printf("Node registered on resolver at height `%d`, tx hash `%s`",
		data.Height, common.HexBytes(data.Tx.Hash()).String())

	resolvers, err := tx.QueryResolver(id)
	if err != nil {
		return nil, err
	}

	isMatch := false
	for _, _resolver := range resolvers {
		if resolver.Equals(_resolver) {
			isMatch = true
		}
	}

	if !isMatch {
		return nil, errors.Errorf("Registered node owner address does not match with resolver address")
	}

	ip, err := PublicIP()
	if err != nil {
		return nil, err
	}

	url := "http://" + resolverIP + "/node/register"
	message := map[string]interface{}{
		"id":   id,
		"ip":   ip,
		"port": port,
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode != 200 {
		log.Fatalln("Error while register on the resolver")
	}

	log.Fatalln("Register node on the resolver completed locally")
	
	return node, nil
}
