package v2ray

import (
	v2raytypes "github.com/sentinel-official/dvpn-node/services/v2ray/types"
	"github.com/sentinel-official/dvpn-node/types"
)

var (
	_ types.Service = (*V2Ray)(nil)
)

type V2Ray struct {
	info []byte
}

func (v *V2Ray) Type() uint64 {
	return v2raytypes.Type
}

func (v *V2Ray) Info() []byte {
	return v.info
}

func (v *V2Ray) Init(home string) error {
	//TODO implement me
	panic("implement me")
}

func (v *V2Ray) Start() error {
	//TODO implement me
	panic("implement me")
}

func (v *V2Ray) Stop() error {
	//TODO implement me
	panic("implement me")
}

func (v *V2Ray) AddPeer(data []byte) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (v *V2Ray) RemovePeer(data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (v *V2Ray) Peers() ([]types.Peer, error) {
	//TODO implement me
	panic("implement me")
}

func (v *V2Ray) PeersCount() int {
	//TODO implement me
	panic("implement me")
}
