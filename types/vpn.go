package types

import (
	hub "github.com/sentinel-official/hub/types"
)

type BaseVPN interface {
	Type() string
	Encryption() string

	Init() error
	Start() error
	Stop() error

	GenerateClientKey(id string) ([]byte, error)
	RevokeClient(id string) error
	DisconnectClient(id string) error
	ClientsList() (map[string]hub.Bandwidth, error)
}
