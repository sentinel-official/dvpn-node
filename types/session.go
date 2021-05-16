package types

import (
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Session struct {
	ID          uint64         `json:"id"`
	Key         string         `json:"key"`
	Address     sdk.AccAddress `json:"address"`
	Available   sdk.Int        `json:"available"`
	Download    int64          `json:"download"`
	Upload      int64          `json:"upload"`
	ConnectedAt time.Time      `json:"connected_at"`
}

type Sessions struct {
	m     map[string]Session
	mutex sync.Mutex
}

func NewSessions() *Sessions {
	return &Sessions{
		m: make(map[string]Session),
	}
}

func (s *Sessions) Put(v Session) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.m[v.Key] = v
}

func (s *Sessions) Get(v string) Session {
	return s.m[v]
}

func (s *Sessions) Delete(v string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.m, v)
}
