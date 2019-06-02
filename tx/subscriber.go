package tx

import (
	"fmt"
	"log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/common"
	coreTypes "github.com/tendermint/tendermint/rpc/core/types"
	rpcTypes "github.com/tendermint/tendermint/rpc/lib/types"
	tmTypes "github.com/tendermint/tendermint/types"

	"github.com/ironman0x7b2/vpn-node/types"
)

type Subscriber struct {
	uri    string
	cdc    *codec.Codec
	conn   *websocket.Conn
	events map[string]chan tmTypes.EventDataTx
}

func NewSubscriber(address string, cdc *codec.Codec) (*Subscriber, error) {
	subscriber := Subscriber{
		uri:    fmt.Sprintf("ws://%s/websocket", address),
		cdc:    cdc,
		events: make(map[string]chan tmTypes.EventDataTx),
	}

	ok := make(chan bool)
	go func() {
		if err := subscriber.ReadTxQuery(ok); err != nil {
			panic(err)
		}
	}()

	<-ok
	return &subscriber, nil
}

func (s *Subscriber) ReadTxQuery(ok chan bool) error {
	log.Printf("Dialing the rpc server `%s`", s.uri)
	conn, _, err := websocket.DefaultDialer.Dial(s.uri, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err := conn.Close(); err != nil {
			panic(err)
		}
	}()

	s.conn = conn
	ok <- true

	var response rpcTypes.RPCResponse
	var result coreTypes.ResultEvent

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		if err := s.cdc.UnmarshalJSON(p, &response); err != nil {
			return err
		}

		if response.Error != nil {
			return response.Error
		}
		if response.Result == nil {
			continue
		}

		if err := s.cdc.UnmarshalJSON(response.Result, &result); err != nil {
			return err
		}

		if result.Data != nil {
			switch data := result.Data.(type) {
			case tmTypes.EventDataTx:
				hash := common.HexBytes(data.Tx.Hash()).String()
				s.events[hash] <- data
				delete(s.events, hash)
			}
		}
	}
}

func (s *Subscriber) WriteTxQuery(hash string, event chan tmTypes.EventDataTx) error {
	if s.conn == nil {
		return errors.Errorf("RPC connection is nil")
	}
	if s.events[hash] != nil {
		return errors.Errorf("Already subscribed to this transaction hash `%s`", hash)
	}

	s.events[hash] = event

	request := types.NewTxRequest(hash)
	if err := s.conn.WriteJSON(request); err != nil {
		delete(s.events, hash)

		return err
	}

	return nil
}
