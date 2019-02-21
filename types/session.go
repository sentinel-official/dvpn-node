package types

import (
	"encoding/base64"
	"sync"
	"time"

	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
	vpnTypes "github.com/ironman0x7b2/sentinel-sdk/x/vpn/types"
	"github.com/pkg/errors"
)

type Session struct {
	*vpnTypes.SessionDetails

	Timeout     *time.Timer
	StopTimeout chan bool

	mutex         sync.Mutex
	bandwidth     sdkTypes.Bandwidth
	nodeOwnerSign []byte
	clientSign    []byte
}

func NewSession(details *vpnTypes.SessionDetails) *Session {
	return &Session{
		SessionDetails: details,
		Timeout:        time.NewTimer(30 * time.Second),
		StopTimeout:    make(chan bool),
	}
}

func (s *Session) VerifyAndSetBandwidthSigns(upload, download int64, _nodeOwnerSign, _clientSign string) error {
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

	bandwidth := sdkTypes.NewBandwidthFromInt64(upload, download)
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

func (s *Session) BandwidthSigns() (sdkTypes.Bandwidth, []byte, []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.bandwidth, s.nodeOwnerSign, s.clientSign
}
