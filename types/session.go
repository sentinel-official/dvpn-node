package types

import (
	"encoding/base64"
	"sync"
	"time"

	csdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/websocket"
	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
	vpnTypes "github.com/ironman0x7b2/sentinel-sdk/x/vpn/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/crypto"
)

type Session struct {
	ID              sdkTypes.ID
	Client          csdkTypes.AccAddress
	NodeOwner       csdkTypes.AccAddress
	ClientPubKey    crypto.PubKey
	NodeOwnerPubKey crypto.PubKey

	bandwidth     sdkTypes.Bandwidth
	nodeOwnerSign []byte
	clientSign    []byte

	Conn            *websocket.Conn
	OutMessages     chan []byte
	StopTimeoutChan chan bool
	Status          string
	mutex           *sync.Mutex
}

func NewSession(details *vpnTypes.SessionDetails) *Session {
	return &Session{
		ID:              details.ID,
		Client:          details.Client,
		NodeOwner:       details.NodeOwner,
		ClientPubKey:    details.ClientPubKey,
		NodeOwnerPubKey: details.NodeOwnerPubKey,

		OutMessages:     make(chan []byte),
		StopTimeoutChan: make(chan bool),
		Status:          vpnTypes.StatusInit,
		mutex:           &sync.Mutex{},
	}
}

func (s *Session) Timeout() <-chan time.Time {
	return time.After(TimeoutSession)
}

func (s *Session) StopTimeout() <-chan bool {
	return s.StopTimeoutChan
}

func (s *Session) VerifyAndSetBandwidthInfo(bandwidth sdkTypes.Bandwidth, _nodeOwnerSign, _clientSign string) error {
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

	if bandwidth.LTE(s.bandwidth) {
		return errors.New("Invalid bandwidth")
	}

	bandwidthSignDataBytes := sdkTypes.NewBandwidthSignData(s.ID, bandwidth, s.NodeOwner, s.Client).GetBytes()
	if !s.ClientPubKey.VerifyBytes(bandwidthSignDataBytes, clientSign) ||
		!s.NodeOwnerPubKey.VerifyBytes(bandwidthSignDataBytes, nodeOwnerSign) {
		return errors.New("Invalid client sign or node owner sign")
	}

	s.bandwidth = bandwidth
	s.nodeOwnerSign = nodeOwnerSign
	s.clientSign = clientSign

	return nil
}

func (s *Session) BandwidthInfo() (sdkTypes.Bandwidth, []byte, []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.bandwidth, s.nodeOwnerSign, s.clientSign
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
