package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"text/template"
	
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	
	"github.com/sentinel-official/dvpn-node/types"
)

// nolint:gochecknoglobals
var (
	appConfigTemplate        *template.Template
	defaultAppConfigTemplate = `
chain_id = "{{ .ChainID }}"
rpc_address = "{{ .RPCAddress }}"
vpn_type = "{{ .VPNType }}"
api_port = {{ .APIPort }}

[account]
name = "{{ .Account.Name }}"

[resolver]
id = "{{ .Resolver.ID }}"
ip = "{{ .Resolver.IP }}"

[node]
id = "{{ .Node.ID }}"
moniker = "{{ .Node.Moniker }}"
description = "{{ .Node.Description }}"
prices_per_gb = "{{ .Node.PricesPerGB }}"
`
)

// nolint:gochecknoinits
func init() {
	var err error
	
	appConfigTemplate, err = template.New("appConfig").Parse(defaultAppConfigTemplate)
	if err != nil {
		panic(err)
	}
}

type AppConfig struct {
	ChainID    string `json:"chain_id"`
	RPCAddress string `json:"rpc_address"`
	VPNType    string `json:"vpn_type"`
	APIPort    uint16 `json:"api_port"`
	
	Account struct {
		Name string `json:"name"`
	} `json:"account"`
	
	Resolver struct {
		ID string `json:"id"`
		IP string `json:"ip"`
	} `json:"resolver"`
	
	Node struct {
		ID          string `json:"id"`
		Moniker     string `json:"moniker"`
		Description string `json:"description"`
		PricesPerGB string `json:"prices_per_gb"`
	} `json:"node"`
}

func NewAppConfig() *AppConfig {
	return &AppConfig{}
}

// nolint:dupl
func (a *AppConfig) LoadFromPath(path string) error {
	if path == "" {
		path = types.DefaultAppConfigFilePath
	}
	
	if _, err := os.Stat(path); err != nil {
		if err := a.SaveToPath(path); err != nil {
			return err
		}
	}
	
	log.Printf("Loading the app configuration from path `%s`", path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	
	if len(data) == 0 {
		*a = AppConfig{}
		return nil
	}
	
	tree, err := toml.LoadBytes(data)
	if err != nil {
		return err
	}
	
	data, err = json.Marshal(tree.ToMap())
	if err != nil {
		return err
	}
	
	return json.Unmarshal(data, a)
}

func (a *AppConfig) SaveToPath(path string) error {
	var buffer bytes.Buffer
	if err := appConfigTemplate.Execute(&buffer, a); err != nil {
		return err
	}
	
	if path == "" {
		path = types.DefaultAppConfigFilePath
	}
	
	return ioutil.WriteFile(path, buffer.Bytes(), os.ModePerm)
}

func (a *AppConfig) Validate() error {
	if a.ChainID == "" {
		return errors.Errorf("Invalid chain_id")
	}
	if a.RPCAddress == "" {
		return errors.Errorf("Invalid rpc_address")
	}
	if a.Resolver.ID == "" {
		return errors.Errorf("Invalid resolver_id")
	}
	if a.Resolver.IP == "" {
		return errors.Errorf("Invalid resolver_ip")
	}
	if a.VPNType == "" {
		return errors.Errorf("Invalid vpn_type")
	}
	if a.Node.Moniker == "" {
		return errors.Errorf("Invalid node.moniker")
	}
	if a.Node.PricesPerGB == "" {
		return errors.Errorf("Invalid node.prices_per_gb")
	}
	
	return nil
}
