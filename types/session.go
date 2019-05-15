package types

import (
	"encoding/base64"
	"sync"
	"time"

	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/websocket"
	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
	"github.com/ironman0x7b2/sentinel-sdk/x/vpn"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/crypto"
)

type Session struct {
	ID              sdkTypes.ID
	NodeOwner       csdkTypes.AccAddress
	NodeOwnerPubKey crypto.PubKey
	Client          csdkTypes.AccAddress
	ClientPubKey    crypto.PubKey
	Status          string

	toProvide     sdkTypes.Bandwidth
	consumed      sdkTypes.Bandwidth
	nodeOwnerSign []byte
	clientSign    []byte

	StopTimeout chan bool
	Conn        *websocket.Conn
	OutMessages chan Msg
	mutex       *sync.Mutex
}

func NewSession(session *vpn.Session) *Session {
	return &Session{
		ID:              session.ID,
		NodeOwner:       session.NodeOwner,
		NodeOwnerPubKey: session.NodeOwnerPubKey,
		Client:          session.Client,
		ClientPubKey:    session.ClientPubKey,
		Status:          session.Status,
		toProvide:       session.BandwidthInfo.ToProvide,
		StopTimeout:     make(chan bool),
		OutMessages:     make(chan Msg, 2),
		mutex:           &sync.Mutex{},
	}
}

func (s Session) Timeout() <-chan time.Time {
	return time.After(SessionTimeout)
}

func (s *Session) StopTimeoutListener() <-chan bool {
	return s.StopTimeout
}

func (s *Session) VerifyAndSetConsumedBandwidth(consumed sdkTypes.Bandwidth, _nodeOwnerSign, _clientSign string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	clientSign, err := base64.StdEncoding.DecodeString(_clientSign)
	if err != nil {
		return err
	}

	nodeOwnerSign, err := base64.StdEncoding.DecodeString(_nodeOwnerSign)
	if err != nil {
		return err
	}

	if consumed.LTE(s.consumed) {
		return errors.Errorf("Invalid consumed bandwidth")
	}

	bandwidthSign := sdkTypes.NewBandwidthSign(s.ID, consumed, s.NodeOwner, s.Client).GetBytes()
	if !s.NodeOwnerPubKey.VerifyBytes(bandwidthSign, nodeOwnerSign) ||
		!s.ClientPubKey.VerifyBytes(bandwidthSign, clientSign) {
		return errors.Errorf("Invalid client sign or node owner sign")
	}

	s.consumed = consumed
	s.nodeOwnerSign = nodeOwnerSign
	s.clientSign = clientSign

	return nil
}

func (s *Session) ConsumedBandwidthInfo() (sdkTypes.Bandwidth, []byte, []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.consumed, s.nodeOwnerSign, s.clientSign
}

type Sessions struct {
	s     map[string]*Session
	mutex *sync.Mutex
}

func NewSessions() Sessions {
	return Sessions{
		s:     make(map[string]*Session),
		mutex: &sync.Mutex{},
	}
}

func (s *Sessions) Set(id string, session *Session) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.s[id] = session
}

func (s *Sessions) Get(id string) *Session {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.s[id]
}

func (s *Sessions) Delete(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.s, id)
}

func (s *Sessions) IDs() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	keys := make([]string, 0, len(s.s))
	for key := range s.s {
		keys = append(keys, key)
	}

	return keys
}
