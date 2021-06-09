package types

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func withTypePrefix(v interface{}) string {
	t := reflect.TypeOf(v).String()
	switch v := v.(type) {
	case string:
		return t + v
	case fmt.Stringer:
		return t + v.String()
	default:
		return ""
	}
}

type Session struct {
	ID          uint64         `json:"id,omitempty"`
	Key         string         `json:"key,omitempty"`
	Address     sdk.AccAddress `json:"address,omitempty"`
	Available   sdk.Int        `json:"available,omitempty"`
	Download    int64          `json:"download,omitempty"`
	Upload      int64          `json:"upload,omitempty"`
	ConnectedAt time.Time      `json:"connected_at,omitempty"`
}

type Sessions struct {
	m     map[string]interface{}
	mutex sync.Mutex
}

func NewSessions() *Sessions {
	return &Sessions{
		m: make(map[string]interface{}),
	}
}

func (s *Sessions) delete(v *Session) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.m, withTypePrefix(v.Key))
	delete(s.m, withTypePrefix(v.Address))
}

func (s *Sessions) Put(v *Session) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.m[withTypePrefix(v.Key)] = v
	s.m[withTypePrefix(v.Address)] = v.Key
}

func (s *Sessions) GetForKey(k string) *Session {
	v, ok := s.m[withTypePrefix(k)]
	if !ok {
		return nil
	}

	return v.(*Session)
}

func (s *Sessions) GetForAddress(k sdk.AccAddress) *Session {
	v, ok := s.m[withTypePrefix(k)]
	if !ok {
		return nil
	}

	return s.GetForKey(v.(string))
}

func (s *Sessions) DeleteForKey(k string) {
	v := s.GetForKey(k)
	if v == nil {
		return
	}

	s.delete(v)
}

func (s *Sessions) DeleteForAddress(k sdk.AccAddress) {
	v := s.GetForAddress(k)
	if v == nil {
		return
	}

	s.delete(v)
}
