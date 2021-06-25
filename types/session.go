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

func (s Session) Empty() bool {
	return s.ID == 0
}

type Sessions struct {
	sync.RWMutex
	m map[string]interface{}
}

func NewSessions() *Sessions {
	return &Sessions{
		m: make(map[string]interface{}),
	}
}

func (s *Sessions) unsafeIsNil(v Session) bool {
	return s.unsafeGetForKey(v.Key).Empty() &&
		s.unsafeGetForAddress(v.Address).Empty()
}

func (s *Sessions) unsafeSet(v Session) {
	s.m[withTypePrefix(v.Key)] = v
	s.m[withTypePrefix(v.Address)] = v.Key
}

func (s *Sessions) unsafeDelete(v Session) {
	delete(s.m, withTypePrefix(v.Key))
	delete(s.m, withTypePrefix(v.Address))
}

func (s *Sessions) unsafeGetForKey(k string) (x Session) {
	v, ok := s.m[withTypePrefix(k)]
	if !ok {
		return x
	}

	return v.(Session)
}

func (s *Sessions) unsafeGetForAddress(k sdk.AccAddress) (x Session) {
	v, ok := s.m[withTypePrefix(k)]
	if !ok {
		return x
	}

	v, ok = s.m[withTypePrefix(v.(string))]
	if !ok {
		return x
	}

	return v.(Session)
}

func (s *Sessions) Set(v Session) {
	s.Lock()
	defer s.Unlock()

	s.unsafeSet(v)
}

func (s *Sessions) Update(v Session) {
	s.Lock()
	defer s.Unlock()

	if s.unsafeIsNil(v) {
		return
	}

	s.unsafeSet(v)
}

func (s *Sessions) GetByKey(k string) Session {
	s.RLock()
	defer s.RUnlock()

	return s.unsafeGetForKey(k)
}

func (s *Sessions) GetByAddress(k sdk.AccAddress) Session {
	s.RLock()
	defer s.RUnlock()

	return s.unsafeGetForAddress(k)
}

func (s *Sessions) DeleteByKey(k string) {
	s.Lock()
	defer s.Unlock()

	if k == "" {
		return
	}

	v := s.unsafeGetForKey(k)
	if v.Empty() {
		return
	}

	s.unsafeDelete(v)
}

func (s *Sessions) DeleteByAddress(k sdk.AccAddress) {
	s.Lock()
	defer s.Unlock()

	if k == nil || k.Empty() {
		return
	}

	v := s.unsafeGetForAddress(k)
	if v.Empty() {
		return
	}

	s.unsafeDelete(v)
}

func (s *Sessions) Len() int {
	s.RLock()
	defer s.RUnlock()

	return len(s.m) / 2
}

func (s *Sessions) Iterate(fn func(v Session) bool) {
	s.RLock()
	defer s.RUnlock()

	for _, v := range s.m {
		v, ok := v.(Session)
		if !ok {
			continue
		}

		if fn(v) {
			break
		}
	}
}
