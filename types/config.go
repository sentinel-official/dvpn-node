package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/pelletier/go-toml"
)

var (
	ct = strings.TrimSpace(`
[chain]
id = "{{ .Chain.ID }}"
rpc_address = "{{ .Chain.RPCAddress }}"
trust_node = {{ .Chain.TrustNode }}
gas = {{ .Chain.Gas }}
gas_adjustment = {{ .Chain.GasAdjustment }}
fees = "{{ .Chain.Fees }}"
gas_prices = "{{ .Chain.GasPrices }}"

[node]
from = "{{ .Node.From }}"
provider = "{{ .Node.Provider }}"
price = "{{ .Node.Price }}"
listen_on = "{{ .Node.ListenOn }}"
category = "{{ .Node.Category }}"
	`)

	t = func() *template.Template {
		t, err := template.New("config").Parse(ct)
		if err != nil {
			panic(err)
		}

		return t
	}()
)

type Config struct {
	Chain struct {
		ID            string  `json:"id"`
		RPCAddress    string  `json:"rpc_address"`
		TrustNode     bool    `json:"trust_node"`
		Gas           uint64  `json:"gas"`
		GasAdjustment float64 `json:"gas_adjustment"`
		Fees          string  `json:"fees"`
		GasPrices     string  `json:"gas_prices"`
	} `json:"chain"`
	Node struct {
		From     string `json:"from"`
		Provider string `json:"provider"`
		Price    string `json:"price"`
		ListenOn string `json:"listen_on"`
		Category string `json:"category"`
	} `json:"node"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) WithDefaultValues() *Config {
	c.Chain.ID = "sentinel-turing-3"
	c.Chain.RPCAddress = "http://127.0.0.1:26657"
	c.Chain.TrustNode = false
	c.Chain.Gas = 1e5
	c.Chain.GasAdjustment = 0
	c.Chain.Fees = ""
	c.Chain.GasPrices = "0.001utsent"

	c.Node.From = ""
	c.Node.Provider = ""
	c.Node.Price = "50utsent"
	c.Node.ListenOn = "0.0.0.0:9656"
	c.Node.Category = "OpenVPN"

	return c
}

func (c *Config) LoadFromPath(path string) error {
	if _, err := os.Stat(path); err != nil {
		cfg := NewConfig().WithDefaultValues()
		if err = cfg.SaveToPath(path); err != nil {
			return err
		}
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		*c = Config{}
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

	return json.Unmarshal(data, c)
}

func (c *Config) SaveToPath(path string) error {
	var buffer bytes.Buffer
	if err := t.Execute(&buffer, c); err != nil {
		return err
	}

	return ioutil.WriteFile(path, buffer.Bytes(), 0600)
}

func (c *Config) String() string {
	var buffer bytes.Buffer
	if err := t.Execute(&buffer, c); err != nil {
		panic(err)
	}

	return buffer.String()
}

func (c *Config) Validate() error {
	if len(c.Chain.ID) == 0 {
		return fmt.Errorf("chain.id is empty")
	}
	if len(c.Chain.RPCAddress) == 0 {
		return fmt.Errorf("chain.rpc_address is empty")
	}
	if len(c.Node.From) == 0 {
		return fmt.Errorf("node.from is empty")
	}
	if len(c.Node.ListenOn) == 0 {
		return fmt.Errorf("node.listen_on is empty")
	}

	return nil
}
