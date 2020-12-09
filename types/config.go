package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/sentinel-official/hub/x/node/types"
)

var (
	ct = strings.TrimSpace(`
[chain]
fees = "{{ .Chain.Fees }}"
gas_adjustment = {{ .Chain.GasAdjustment }}
gas = {{ .Chain.Gas }}
gas_prices = "{{ .Chain.GasPrices }}"
id = "{{ .Chain.ID }}"
rpc_address = "{{ .Chain.RPCAddress }}"
trust_node = {{ .Chain.TrustNode }}

[handshake]
enable = {{ .Handshake.Enable }}
peers = {{ .Handshake.Peers }}

[node]
from = "{{ .Node.From }}"
interval_sessions = {{ .Node.IntervalSessions }}
interval_status = {{ .Node.IntervalStatus }}
listen_on = "{{ .Node.ListenOn }}"
moniker = "{{ .Node.Moniker }}"
price = "{{ .Node.Price }}"
provider = "{{ .Node.Provider }}"
remote_url = "{{ .Node.RemoteURL }}"
type = "{{ .Node.Type }}"
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
		Fees          string  `json:"fees"`
		GasAdjustment float64 `json:"gas_adjustment"`
		GasPrices     string  `json:"gas_prices"`
		Gas           uint64  `json:"gas"`
		ID            string  `json:"id"`
		RPCAddress    string  `json:"rpc_address"`
		TrustNode     bool    `json:"trust_node"`
	} `json:"chain"`
	Handshake struct {
		Enable bool   `json:"enable"`
		Peers  uint64 `json:"peers"`
	}
	Node struct {
		From             string `json:"from"`
		IntervalSessions int64  `json:"interval_sessions"`
		IntervalStatus   int64  `json:"interval_status"`
		ListenOn         string `json:"listen_on"`
		Moniker          string `json:"moniker"`
		Price            string `json:"price"`
		Provider         string `json:"provider"`
		RemoteURL        string `json:"remote_url"`
		Type             string `json:"type"`
	} `json:"node"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) WithDefaultValues() *Config {
	c.Chain.Fees = ""
	c.Chain.Gas = 1e5
	c.Chain.GasAdjustment = 0
	c.Chain.GasPrices = "0.01tsent"
	c.Chain.ID = "sentinel-turing-3a"
	c.Chain.RPCAddress = "https://rpc.turing.sentinel.co:443"
	c.Chain.TrustNode = false
	c.Handshake.Enable = true
	c.Handshake.Peers = 8
	c.Node.From = ""
	c.Node.IntervalSessions = 8 * time.Minute.Nanoseconds()
	c.Node.IntervalStatus = 4 * time.Minute.Nanoseconds()
	c.Node.ListenOn = "127.0.0.1:8585"
	c.Node.Moniker = ""
	c.Node.Price = "50tsent"
	c.Node.Provider = ""
	c.Node.RemoteURL = ""
	c.Node.Type = types.CategoryWireGuard.String()

	return c
}

func (c *Config) LoadFromPath(path string) error {
	if _, err := os.Stat(path); err != nil {
		cfg := NewConfig().WithDefaultValues()
		if err := cfg.SaveToPath(path); err != nil {
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
	if c.Chain.ID == "" {
		return fmt.Errorf("invalid chain.id")
	}
	if c.Chain.RPCAddress == "" {
		return fmt.Errorf("invalid chain.rpc_address")
	}
	if c.Handshake.Peers == 0 {
		return fmt.Errorf("invalid handshake.peers")
	}
	if c.Node.From == "" {
		return fmt.Errorf("invalid node.from")
	}
	if c.Node.IntervalSessions <= 0 {
		return fmt.Errorf("invalid node.interval_sessions")
	}
	if c.Node.IntervalStatus <= 0 {
		return fmt.Errorf("invalid node.interval_status")
	}
	if c.Node.ListenOn == "" {
		return fmt.Errorf("invalid node.listen_on")
	}
	if (c.Node.Provider != "" && c.Node.Price != "") ||
		(c.Node.Provider == "" && c.Node.Price == "") {
		return fmt.Errorf("invalid node.provider or node.price")
	}
	if c.Node.RemoteURL == "" {
		return fmt.Errorf("invalid node.remote_url")
	}
	if c.Node.Type == "" {
		return fmt.Errorf("invalid node.type")
	}

	return nil
}
