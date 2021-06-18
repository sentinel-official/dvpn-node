package types

import (
	"encoding/json"
	"io/ioutil"

	"github.com/sentinel-official/dvpn-node/services/v2ray/protocols"
)

var (
	SupportedProtocols = [...]Protocol{protocols.VMess{}, protocols.Trojan{}}
)

type Config struct {
	HomeDir           string
	APIPort           uint16
	VMessPort         uint16
	TrojanPort        uint16
	ProtocolSelection uint16
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) LoadFromPath(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, c)
}
