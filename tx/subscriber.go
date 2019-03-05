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
	nodeURI  string
	cdc      *codec.Codec
	conn     *websocket.Conn
	channels map[string]chan tmTypes.EventDataTx
}

func NewSubscriber(liteClientURI string, cdc *codec.Codec) (*Subscriber, error) {
	subscriber := Subscriber{
		nodeURI:  fmt.Sprintf("ws://%s/websocket", liteClientURI),
		cdc:      cdc,
		channels: make(map[string]chan tmTypes.EventDataTx),
	}

	ok := make(chan bool)
	defer close(ok)

	go func() {
		if err := subscriber.ReadTxQuery(ok); err != nil {
			panic(err)
		}
	}()

	<-ok

	return &subscriber, nil
}

func (s *Subscriber) ReadTxQuery(ok chan bool) error {
	log.Printf("Dialing the node with URI `%s`", s.nodeURI)
	conn, _, err := websocket.DefaultDialer.Dial(s.nodeURI, nil)
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
				txHash := common.HexBytes(data.Tx.Hash()).String()
				s.channels[txHash] <- data
				delete(s.channels, txHash)
			}
		}
	}
}

func (s *Subscriber) WriteTxQuery(txHash string, channel chan tmTypes.EventDataTx) error {
	if s.conn == nil {
		return errors.New("Connection is nil")
	}
	if s.channels[txHash] != nil {
		return errors.New("Already subscribed")
	}

	s.channels[txHash] = channel

	body := types.NewTxSubscriberRPCRequest(txHash)
	if err := s.conn.WriteJSON(body); err != nil {
		delete(s.channels, txHash)

		return err
	}

	return nil
}
