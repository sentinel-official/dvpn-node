package types

import (
	"github.com/sentinel-official/hub/x/node"
)

type Service interface {
	Type() node.Category
	Initialize(home string) error
	Start() error
	Stop() error
	AddPeer([]byte) ([]byte, error)
	RemovePeer([]byte) error
	Peers() ([]Peer, error)
	PeersCount() int
}

type Peer struct {
	Identity string `json:"identity"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}
