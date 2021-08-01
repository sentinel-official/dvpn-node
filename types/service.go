package types

import (
	tmlog "github.com/tendermint/tendermint/libs/log"
)

type Service interface {
	Type() uint64
	Info() []byte
	WithLogger(logger tmlog.Logger) Service
	PreInit(home string) error
	Init(home string) error
	PostInit(home string) error
	PreUp() error
	Up() error
	PostUp() error
	PreDown() error
	Down() error
	PostDown() error
	Start() error
	Stop() error
	AddPeer(data []byte) ([]byte, error)
	RemovePeer(data []byte) error
	Peers() ([]Peer, error)
	PeersLen() int
}

type Peer struct {
	Key      string `json:"key"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}
