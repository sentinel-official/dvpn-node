package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/ironman0x7b2/vpn-node/types"
)

type OpenVPNConfig struct {
	Port             uint16 `json:"port"`
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

	log.Printf("Loading the OpenVPN configuration from path `%s`", path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		log.Println("Found an empty OpenVPN configuration")
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

	log.Printf("Saving the OpenVPN configuration to path `%s`", path)
	return ioutil.WriteFile(path, data, os.ModePerm)
}
