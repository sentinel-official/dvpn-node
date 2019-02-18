package tx

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/rpc/lib/types"
	"github.com/tendermint/tendermint/types"
)

type Subscriber struct {
	nodeURI  string
	cdc      *amino.Codec
	conn     *websocket.Conn
	channels map[string]chan types.EventDataTx
}

func NewSubscriber(nodeURI string, cdc *amino.Codec) (*Subscriber, error) {
	subscriber := Subscriber{
		nodeURI:  nodeURI,
		cdc:      cdc,
		channels: make(map[string]chan types.EventDataTx),
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

	var rpcResponse rpctypes.RPCResponse
	var resultEvent core_types.ResultEvent

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		if err := s.cdc.UnmarshalJSON(p, &rpcResponse); err != nil {
			return err
		}

		if rpcResponse.Error != nil {
			return fmt.Errorf(rpcResponse.Error.Error())
		}
		if rpcResponse.Result == nil {
			continue
		}

		if err := s.cdc.UnmarshalJSON(rpcResponse.Result, &resultEvent); err != nil {
			return err
		}

		if resultEvent.Data != nil {
			switch data := resultEvent.Data.(type) {
			case types.EventDataTx:
				txHash := common.HexBytes(data.Tx.Hash()).String()
				s.channels[txHash] <- data
				close(s.channels[txHash])
				delete(s.channels, txHash)
			}
		}
	}
}

func (s *Subscriber) WriteTxQuery(txHash string, channel chan types.EventDataTx) error {
	if s.conn == nil {
		return fmt.Errorf("connection is nil")
	}
	if s.channels[txHash] != nil {
		return fmt.Errorf("already subscribed")
	}

	s.channels[txHash] = channel

	body := NewTxSubscriberRPCRequest(txHash)
	if err := s.conn.WriteJSON(body); err != nil {
		delete(s.channels, txHash)

		return err
	}

	return nil
}
