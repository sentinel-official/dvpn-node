package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/ironman0x7b2/vpn-node/types"
)

type OpenVPNConfig struct {
	Port             uint16 `json:"port"`
	ManagementPort   uint16 `json:"management_port"`
	IPAddress        string `json:"ip_address"`
	Protocol         string `json:"protocol"`
	EncryptionMethod string `json:"encryption_method"`
}

func NewOpenVPNConfig() *OpenVPNConfig {
	return &OpenVPNConfig{}
}

func (o *OpenVPNConfig) LoadFromPath(path string) error {
	if len(path) == 0 {
		path = types.DefaultOpenVPNConfigFilePath
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		data, err = json.Marshal(OpenVPNConfig{})
		if err != nil {
			return err
		}
	}

	return json.Unmarshal(data, o)
}

func (o OpenVPNConfig) SaveToPath(path string) error {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err
	}

	if len(path) == 0 {
		path = types.DefaultOpenVPNConfigFilePath
	}

	return ioutil.WriteFile(path, data, os.ModePerm)
}
