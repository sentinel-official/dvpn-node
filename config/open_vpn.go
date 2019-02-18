package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type OpenVPNConfig struct {
	IPAddress        string `json:"ip_address"`
	Port             uint32 `json:"port"`
	Protocol         string `json:"protocol"`
	EncryptionMethod string `json:"encryption_method"`
	ManagementPort   uint32 `json:"management_port"`
}

func NewOpenVPNConfig() *OpenVPNConfig {
	return &OpenVPNConfig{}
}

func (o *OpenVPNConfig) LoadFromPath(path string) error {
	if len(path) == 0 {
		path = DefaultOpenVPNConfigFilePath
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
		path = DefaultOpenVPNConfigFilePath
	}

	return ioutil.WriteFile(path, data, os.ModePerm)
}
