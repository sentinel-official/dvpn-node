package types

import (
	"encoding/base64"
	"sync"

	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
	vpnTypes "github.com/ironman0x7b2/sentinel-sdk/x/vpn/types"
	"github.com/pkg/errors"
)

type Session struct {
	*vpnTypes.SessionDetails
	StopTimeout chan bool
	mutex       *sync.Mutex
}

func NewSession(details *vpnTypes.SessionDetails) *Session {
	return &Session{
		SessionDetails: details,
		StopTimeout:    make(chan bool),
		mutex:          &sync.Mutex{},
	}
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

	if bandwidth.LTE(s.Bandwidth.Consumed) {
		return errors.New("Invalid bandwidth")
	}

	bandwidthSignDataBytes := sdkTypes.NewBandwidthSignData(s.ID, bandwidth, s.NodeOwner, s.Client).GetBytes()
	if !s.ClientPubKey.VerifyBytes(bandwidthSignDataBytes, clientSign) ||
		!s.NodeOwnerPubKey.VerifyBytes(bandwidthSignDataBytes, nodeOwnerSign) {
		return errors.New("Invalid client sign or node owner sign")
	}

	s.Bandwidth.Consumed = bandwidth
	s.Bandwidth.NodeOwnerSign = nodeOwnerSign
	s.Bandwidth.ClientSign = clientSign

	return nil
}

func (s *Session) BandwidthInfo() (sdkTypes.Bandwidth, []byte, []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	b := s.Bandwidth

	return b.Consumed, b.NodeOwnerSign, b.ClientSign
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
