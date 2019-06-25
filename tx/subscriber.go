package tx

import (
	"fmt"
	"log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/common"
	core "github.com/tendermint/tendermint/rpc/core/types"
	rpc "github.com/tendermint/tendermint/rpc/lib/types"
	tm "github.com/tendermint/tendermint/types"

	"github.com/sentinel-official/dvpn-node/types"
)

type Subscriber struct {
	uri    string
	cdc    *codec.Codec
	conn   *websocket.Conn
	events map[string]chan tm.EventDataTx
}

func NewSubscriber(address string, cdc *codec.Codec) (*Subscriber, error) {
	uri := fmt.Sprintf("ws://%s/websocket", address)

	log.Printf("Dialing the rpc server `%s`", uri)
	conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
	if err != nil {
		return nil, err
	}

	return &Subscriber{
		uri:    uri,
		cdc:    cdc,
		conn:   conn,
		events: make(map[string]chan tm.EventDataTx),
	}, nil
}

// nolint:gocyclo
func (s *Subscriber) ReadTxQuery() error {
	defer func() {
		if err := s.conn.Close(); err != nil {
			panic(err)
		}
	}()

	var response rpc.RPCResponse
	var result core.ResultEvent

	for {
		_, p, err := s.conn.ReadMessage()
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

		if result.Data == nil {
			continue
		}

		data, ok := result.Data.(tm.EventDataTx)
		if !ok {
			continue
		}

		hash := common.HexBytes(data.Tx.Hash()).String()
		s.events[hash] <- data
		delete(s.events, hash)

		request := types.NewTxUnSubscribeRequest(hash)
		if err := s.conn.WriteJSON(request); err != nil {
			return err
		}
	}
}

func (s *Subscriber) WriteTxQuery(hash string, event chan tm.EventDataTx) error {
	if s.conn == nil {
		return errors.Errorf("RPC connection is nil")
	}
	if s.events[hash] != nil {
		return errors.Errorf("Already subscribed to this transaction hash `%s`", hash)
	}

	s.events[hash] = event

	request := types.NewTxSubscribeRequest(hash)
	if err := s.conn.WriteJSON(request); err != nil {
		delete(s.events, hash)

		return err
	}

	return nil
}
