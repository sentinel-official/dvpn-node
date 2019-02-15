package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type AppConfig struct {
	OwnerAccount struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	} `json:"owner_account"`
	Node struct {
		ID           string `json:"id"`
		AmountToLock string `json:"amount_to_lock"`
		PricesPerGB  string `json:"prices_per_gb"`
		Description  string `json:"description"`
		Status       string `json:"status"`
		API          struct {
			Address string `json:"address"`
			Port    uint32 `json:"port"`
		} `json:"api"`
	} `json:"node"`
	VPN struct {
		Type             string `json:"type"`
		IPAddress        string `json:"ip_address"`
		Port             uint32 `json:"port"`
		Protocol         string `json:"protocol"`
		EncryptionMethod string `json:"encryption_method"`
		ManagementPort   uint32 `json:"management_port"`
	} `json:"vpn"`
	LiteClientURI string `json:"lite_client_uri"`
	ChainID       string `json:"chain_id"`
}

func NewAppConfig() *AppConfig {
	return &AppConfig{}
}

func (c *AppConfig) LoadFromPath(path string) error {
	if len(path) == 0 {
		path = DefaultAppConfigFilePath
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	data, err := ioutil.ReadAll(file)
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
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	if len(path) == 0 {
		path = DefaultAppConfigFilePath
	}

	return ioutil.WriteFile(path, data, os.ModePerm)
}
