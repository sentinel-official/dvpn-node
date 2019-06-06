package types

import (
	sdk "github.com/ironman0x7b2/sentinel-sdk/types"
)

type BaseVPN interface {
	Type() string
	Encryption() string

	Init() error
	Start() error
	Stop() error

	GenerateClientKey(id string) ([]byte, error)
	DisconnectClient(id string) error
	ClientsList() (map[string]sdk.Bandwidth, error)
}
