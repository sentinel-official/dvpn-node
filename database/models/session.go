package models

import (
	"encoding/base64"
	"fmt"
	"time"

	"cosmossdk.io/math"
	cosmossdk "github.com/cosmos/cosmos-sdk/types"
	sentinelhub "github.com/sentinel-official/sentinelhub/v12/types"
	"github.com/sentinel-official/sentinelhub/v12/x/session/types/v3"
)

// Session represents a session record in the database. Per-peer data lives in the
// child session_peers table; the parent row holds only session-wide fields.
type Session struct {
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"` // Timestamp when the record was created
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"` // Timestamp when the record was last updated

	NodeAddr string `gorm:"column:node_addr;index:idx_node_addr;not null"` // Address of the node associated with the session

	AccAddr     string        `gorm:"column:acc_addr;index:idx_acc_addr;not null"` // Account address, cannot be null
	ID          uint64        `gorm:"column:id;not null;primaryKey"`               // Unique identifier for the session
	MaxBytes    string        `gorm:"column:max_bytes;not null"`                   // Maximum bytes represented as a string
	MaxDuration time.Duration `gorm:"column:max_duration;not null"`                // Maximum allowed duration for the session in nanoseconds
	Signature   string        `gorm:"column:signature;not null"`                   // Signature associated with the session (empty allowed)

	Peers []SessionPeer `gorm:"foreignKey:SessionID;references:ID;constraint:OnDelete:CASCADE"` // Child peers, one per protocol
}

// NewSession creates and returns a new instance of the Session struct with default values.
func NewSession() *Session {
	return &Session{}
}

// WithAccAddr sets the AccAddr field and returns the updated Session instance.
func (s *Session) WithAccAddr(v cosmossdk.AccAddress) *Session {
	s.AccAddr = v.String()

	return s
}

// WithID sets the ID field and returns the updated Session instance.
func (s *Session) WithID(v uint64) *Session {
	s.ID = v

	return s
}

// WithMaxBytes sets the MaxBytes field from math.Int and returns the updated Session instance.
func (s *Session) WithMaxBytes(v math.Int) *Session {
	s.MaxBytes = v.String()

	return s
}

// WithMaxDuration sets the MaxDuration field from time.Duration and returns the updated Session instance.
func (s *Session) WithMaxDuration(v time.Duration) *Session {
	s.MaxDuration = v

	return s
}

// WithNodeAddr sets the NodeAddr field and returns the updated Session instance.
func (s *Session) WithNodeAddr(v sentinelhub.NodeAddress) *Session {
	s.NodeAddr = v.String()

	return s
}

// WithSignature sets the Signature field and returns the updated Session instance.
func (s *Session) WithSignature(v []byte) *Session {
	s.Signature = base64.StdEncoding.EncodeToString(v)

	return s
}

// GetAccAddr returns the AccAddr field as cosmossdk.AccAddress.
func (s *Session) GetAccAddr() cosmossdk.AccAddress {
	addr, err := cosmossdk.AccAddressFromBech32(s.AccAddr)
	if err != nil {
		panic(fmt.Errorf("decoding Bech32 account addr %q: %w", s.AccAddr, err))
	}

	return addr
}

// GetID returns the ID field.
func (s *Session) GetID() uint64 {
	return s.ID
}

// GetMaxBytes returns the MaxBytes field as math.Int.
func (s *Session) GetMaxBytes() math.Int {
	v, ok := math.NewIntFromString(s.MaxBytes)
	if !ok {
		panic(fmt.Errorf("parsing max_bytes %q", s.MaxBytes))
	}

	return v
}

// GetMaxDuration returns the MaxDuration field as time.Duration.
func (s *Session) GetMaxDuration() time.Duration {
	return s.MaxDuration
}

// GetNodeAddr returns the NodeAddr field as sentinelhub.NodeAddress.
func (s *Session) GetNodeAddr() sentinelhub.NodeAddress {
	addr, err := sentinelhub.NodeAddressFromBech32(s.NodeAddr)
	if err != nil {
		panic(fmt.Errorf("decoding Bech32 node addr %q: %w", s.NodeAddr, err))
	}

	return addr
}

// GetSignature returns the Signature field as a byte slice.
func (s *Session) GetSignature() []byte {
	if s.Signature == "" {
		return nil
	}

	buf, err := base64.StdEncoding.DecodeString(s.Signature)
	if err != nil {
		panic(fmt.Errorf("decoding Base64 signature %q: %w", s.Signature, err))
	}

	return buf
}

// MsgUpdateSessionRequest builds an update message from aggregated peer usage:
// downloadBytes/uploadBytes are summed bytes, duration is MAX across peers.
func (s *Session) MsgUpdateSessionRequest(downloadBytes, uploadBytes math.Int, duration time.Duration) *v3.MsgUpdateSessionRequest {
	return v3.NewMsgUpdateSessionRequest(
		s.GetNodeAddr(),
		s.GetID(),
		downloadBytes,
		uploadBytes,
		duration,
		s.GetSignature(),
	)
}
