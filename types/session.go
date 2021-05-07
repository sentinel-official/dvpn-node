package types

import (
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Session struct {
	Address      sdk.AccAddress `json:"address"`
	ConnectedAt  time.Time      `json:"connected_at"`
	Download     int64          `json:"download"`
	Duration     time.Duration  `json:"duration"`
	Identity     string         `json:"identity"`
	Subscription uint64         `json:"subscription"`
	Upload       int64          `json:"upload"`
}

type Sessions struct {
	m     map[string]Session
	mutex sync.Mutex
}

func NewSessions() *Sessions {
	return &Sessions{
		m:     make(map[string]Session),
		mutex: sync.Mutex{},
	}
}

func (s *Sessions) Put(v Session) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.m[v.Identity] = v
}

func (s *Sessions) Get(v string) Session {
	return s.m[v]
}

func (s *Sessions) Delete(v string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.m, v)
}
