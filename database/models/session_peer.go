package models

import (
	"encoding/base64"
	"fmt"
	"time"

	"cosmossdk.io/math"
	sentinelsdk "github.com/sentinel-official/sentinel-go-sdk/types"
	"gorm.io/gorm"
)

// SessionPeer represents a per-protocol peer entry belonging to a session.
type SessionPeer struct {
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"` // Timestamp when the record was created
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"` // Timestamp when the record was last updated

	SessionID   uint64 `gorm:"column:session_id;primaryKey"`                                                                                                    // FK → sessions.id; composite PK with service_type
	ServiceType string `gorm:"column:service_type;primaryKey;not null;uniqueIndex:uix_service_peer,priority:1;uniqueIndex:uix_service_peer_request,priority:1"` // VPN protocol type string; composite PK with session_id

	PeerID      string `gorm:"column:peer_id;not null;uniqueIndex:uix_service_peer,priority:2"`              // Peer identifier (e.g. WireGuard public key); unique per service
	PeerRequest string `gorm:"column:peer_request;not null;uniqueIndex:uix_service_peer_request,priority:2"` // Base64-encoded protocol peer request; unique per service

	PeerMetadata string `gorm:"column:peer_metadata;not null"` // Base64-encoded peer metadata

	RxBytes string `gorm:"column:rx_bytes;not null"` // Received bytes represented as a string
	TxBytes string `gorm:"column:tx_bytes;not null"` // Transmitted bytes represented as a string

	Duration time.Duration `gorm:"column:duration;not null"` // Elapsed duration snapshot (nanoseconds); updated by BeforeUpdate hook
}

// NewSessionPeer creates and returns a new SessionPeer with default zero values.
func NewSessionPeer() *SessionPeer {
	return &SessionPeer{}
}

// WithSessionID sets the SessionID field and returns the updated SessionPeer.
func (s *SessionPeer) WithSessionID(v uint64) *SessionPeer {
	s.SessionID = v
	return s
}

// WithServiceType sets the ServiceType field from a sentinelsdk.ServiceType and returns the updated SessionPeer.
func (s *SessionPeer) WithServiceType(v sentinelsdk.ServiceType) *SessionPeer {
	s.ServiceType = v.String()
	return s
}

// WithPeerID sets the PeerID field and returns the updated SessionPeer.
func (s *SessionPeer) WithPeerID(v string) *SessionPeer {
	s.PeerID = v
	return s
}

// WithPeerRequest sets the PeerRequest field from a raw byte slice (base64-encoded) and returns the updated SessionPeer.
func (s *SessionPeer) WithPeerRequest(v []byte) *SessionPeer {
	s.PeerRequest = base64.StdEncoding.EncodeToString(v)
	return s
}

// WithPeerMetadata sets the PeerMetadata field from a raw byte slice (base64-encoded) and returns the updated SessionPeer.
func (s *SessionPeer) WithPeerMetadata(v []byte) *SessionPeer {
	s.PeerMetadata = base64.StdEncoding.EncodeToString(v)
	return s
}

// WithRxBytes sets the RxBytes field from math.Int and returns the updated SessionPeer.
func (s *SessionPeer) WithRxBytes(v math.Int) *SessionPeer {
	s.RxBytes = v.String()
	return s
}

// WithTxBytes sets the TxBytes field from math.Int and returns the updated SessionPeer.
func (s *SessionPeer) WithTxBytes(v math.Int) *SessionPeer {
	s.TxBytes = v.String()
	return s
}

// WithDuration sets the Duration field and returns the updated SessionPeer.
func (s *SessionPeer) WithDuration(v time.Duration) *SessionPeer {
	s.Duration = v
	return s
}

// GetSessionID returns the SessionID field.
func (s *SessionPeer) GetSessionID() uint64 {
	return s.SessionID
}

// GetServiceType returns the ServiceType field as a sentinelsdk.ServiceType.
func (s *SessionPeer) GetServiceType() sentinelsdk.ServiceType {
	return sentinelsdk.ServiceTypeFromString(s.ServiceType)
}

// GetPeerID returns the PeerID field.
func (s *SessionPeer) GetPeerID() string {
	return s.PeerID
}

// GetPeerRequest returns the PeerRequest field as a decoded byte slice.
func (s *SessionPeer) GetPeerRequest() []byte {
	buf, err := base64.StdEncoding.DecodeString(s.PeerRequest)
	if err != nil {
		panic(fmt.Errorf("decoding Base64 peer request %q: %w", s.PeerRequest, err))
	}
	return buf
}

// GetPeerMetadata returns the PeerMetadata field as a decoded byte slice.
func (s *SessionPeer) GetPeerMetadata() []byte {
	buf, err := base64.StdEncoding.DecodeString(s.PeerMetadata)
	if err != nil {
		panic(fmt.Errorf("decoding Base64 peer metadata %q: %w", s.PeerMetadata, err))
	}
	return buf
}

// GetRxBytes returns the RxBytes field as math.Int.
func (s *SessionPeer) GetRxBytes() math.Int {
	v, ok := math.NewIntFromString(s.RxBytes)
	if !ok {
		panic(fmt.Errorf("parsing rx_bytes %q", s.RxBytes))
	}
	return v
}

// GetTxBytes returns the TxBytes field as math.Int.
func (s *SessionPeer) GetTxBytes() math.Int {
	v, ok := math.NewIntFromString(s.TxBytes)
	if !ok {
		panic(fmt.Errorf("parsing tx_bytes %q", s.TxBytes))
	}
	return v
}

// GetTotalBytes returns the sum of rx and tx bytes as math.Int.
func (s *SessionPeer) GetTotalBytes() math.Int {
	return s.GetRxBytes().Add(s.GetTxBytes())
}

// GetDuration returns the Duration field.
func (s *SessionPeer) GetDuration() time.Duration {
	return s.Duration
}

// BeforeUpdate is a GORM hook that snapshots the elapsed duration when rx_bytes or tx_bytes change.
func (s *SessionPeer) BeforeUpdate(db *gorm.DB) error {
	if s.SessionID == 0 {
		return nil
	}

	if db.Statement.Changed("rx_bytes", "tx_bytes") {
		duration := time.Since(s.CreatedAt).Nanoseconds()
		db.Statement.SetColumn("duration", duration)
	}

	return nil
}
