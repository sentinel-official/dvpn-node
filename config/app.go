package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/ironman0x7b2/vpn-node/types"
)

type AppConfig struct {
	Owner struct {
		Name     string `json:"name"`
		Address  string `json:"address"`
		Password string `json:"password,omitempty"`
	} `json:"owner"`
	Node struct {
		ID           string `json:"id"`
		AmountToLock string `json:"amount_to_lock"`
		PricesPerGB  string `json:"prices_per_gb"`
		Description  string `json:"description"`
		APIPort      uint16 `json:"api_port"`
	} `json:"node"`
	VPNType       string `json:"vpn_type"`
	LiteClientURI string `json:"lite_client_uri"`
	ChainID       string `json:"chain_id"`
}

func NewAppConfig() *AppConfig {
	return &AppConfig{}
}

func (c *AppConfig) LoadFromPath(path string) error {
	if len(path) == 0 {
		path = types.DefaultAppConfigFilePath
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		data, err = json.Marshal(AppConfig{})
		if err != nil {
			return err
		}
	}

	return json.Unmarshal(data, c)
}

func (c AppConfig) SaveToPath(path string) error {
	_c := c
	_c.Owner.Password = ""

	data, err := json.MarshalIndent(_c, "", "  ")
	if err != nil {
		return err
	}

	if len(path) == 0 {
		path = types.DefaultAppConfigFilePath
	}

	return ioutil.WriteFile(path, data, os.ModePerm)
}
