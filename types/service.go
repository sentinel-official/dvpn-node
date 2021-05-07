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
	PeersCount() int
}

type Peer struct {
	Identity string `json:"identity"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}
