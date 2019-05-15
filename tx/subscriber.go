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
	rpcURI string
	cdc    *codec.Codec
	conn   *websocket.Conn
	events map[string]chan tmTypes.EventDataTx
}

func NewSubscriber(rpcServerAddress string, cdc *codec.Codec) (*Subscriber, error) {
	subscriber := Subscriber{
		rpcURI: fmt.Sprintf("ws://%s/websocket", rpcServerAddress),
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
	log.Printf("Dialing the rpc server `%s`", s.rpcURI)
	conn, _, err := websocket.DefaultDialer.Dial(s.rpcURI, nil)
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

	var rpcResponse rpcTypes.RPCResponse
	var resultEvent coreTypes.ResultEvent

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		if err := s.cdc.UnmarshalJSON(p, &rpcResponse); err != nil {
			return err
		}

		if rpcResponse.Error != nil {
			return rpcResponse.Error
		}
		if rpcResponse.Result == nil {
			continue
		}

		if err := s.cdc.UnmarshalJSON(rpcResponse.Result, &resultEvent); err != nil {
			return err
		}

		if resultEvent.Data != nil {
			switch data := resultEvent.Data.(type) {
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
		return errors.Errorf("Already subscribed to this transaction hash `$s`", hash)
	}

	s.events[hash] = event

	body := types.NewTxSubscriberRPCRequest(hash)
	if err := s.conn.WriteJSON(body); err != nil {
		delete(s.events, hash)
		return err
	}

	return nil
}
