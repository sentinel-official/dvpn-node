package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/ironman0x7b2/vpn-node/types"
)

type AppConfig struct {
	ChainID          string `json:"chain_id"`
	RPCServerAddress string `json:"rpc_server_address"`
	Resolver         string `json:"resolver"`
	Account          struct {
		Name     string `json:"name"`
		Password string `json:"password,omitempty"`
	} `json:"account"`
	APIPort uint16 `json:"api_port"`
	VPNType string `json:"vpn_type"`
	Node    struct {
		ID          string `json:"id"`
		Moniker     string `json:"moniker"`
		Description string `json:"description"`
		PricesPerGB string `json:"prices_per_gb"`
	} `json:"node"`
}

func NewAppConfig() *AppConfig {
	return &AppConfig{}
}

func (c *AppConfig) LoadFromPath(path string) error {
	if len(path) == 0 {
		path = types.DefaultAppConfigFilePath
	}

	log.Printf("Loading the app configuration from path `%s`", path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		log.Println("Found an empty app configuration")
		data, err = json.Marshal(AppConfig{})
		if err != nil {
			return err
		}
	}

	return json.Unmarshal(data, c)
}

func (c AppConfig) SaveToPath(path string) error {
	c.Account.Password = ""

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	if len(path) == 0 {
		path = types.DefaultAppConfigFilePath
	}

	log.Printf("Saving the app configuration to path `%s`", path)
	return ioutil.WriteFile(path, data, os.ModePerm)
}
