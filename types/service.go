package types

type Service interface {
	Type() uint64
	Info() []byte
	Init(home string) error
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
