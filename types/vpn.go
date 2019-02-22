package types

import (
	"sort"
)

type BaseVPN interface {
	Type() string
	Encryption() string

	Init() error
	Start() error
	Stop() error
	Wait(chan error)

	GenerateClientKey(id string) ([]byte, error)
	DisconnectClient(id string) error
	ClientList() (VPNClients, error)
}

type VPNClient struct {
	ID       string `json:"id"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
}

func NewVPNClient(id string, upload, download int64) VPNClient {
	return VPNClient{
		ID:       id,
		Upload:   upload,
		Download: download,
	}
}

type VPNClients []VPNClient

func NewVPNClients() VPNClients {
	return VPNClients{}
}

func (v VPNClients) Append(clients ...VPNClient) VPNClients { return append(v, clients...) }
func (v VPNClients) Len() int                               { return len(v) }
func (v VPNClients) Less(i, j int) bool                     { return v[i].ID < v[j].ID }
func (v VPNClients) Swap(i, j int)                          { v[i], v[j] = v[j], v[i] }

func (v VPNClients) Sort() VPNClients {
	sort.Sort(v)
	return v
}

func (v VPNClients) Remove(index int) VPNClients {
	return NewVPNClients().Append(v[:index]...).Append(v[index+1:]...)
}
