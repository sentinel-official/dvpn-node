package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"
	"time"

	"github.com/pelletier/go-toml"

	wgtypes "github.com/sentinel-official/dvpn-node/services/wireguard/types"
)

var (
	ct = strings.TrimSpace(`
[chain]
gas_adjustment = {{ .Chain.GasAdjustment }}
gas = {{ .Chain.Gas }}
gas_prices = "{{ .Chain.GasPrices }}"
id = "{{ .Chain.ID }}"
rpc_address = "{{ .Chain.RPCAddress }}"

[handshake]
enable = {{ .Handshake.Enable }}
peers = {{ .Handshake.Peers }}

[keyring]
backend = "{{ .Keyring.Backend }}"

[node]
from = "{{ .Node.From }}"
interval_sessions = {{ .Node.IntervalSessions }}
interval_status = {{ .Node.IntervalStatus }}
listen_on = "{{ .Node.ListenOn }}"
moniker = "{{ .Node.Moniker }}"
price = "{{ .Node.Price }}"
provider = "{{ .Node.Provider }}"
remote_url = "{{ .Node.RemoteURL }}"
type = {{ .Node.Type }}
	`)

	t = func() *template.Template {
		t, err := template.New("").Parse(ct)
		if err != nil {
			panic(err)
		}

		return t
	}()
)

type Config struct {
	Chain struct {
		GasAdjustment float64 `json:"gas_adjustment"`
		GasPrices     string  `json:"gas_prices"`
		Gas           uint64  `json:"gas"`
		ID            string  `json:"id"`
		RPCAddress    string  `json:"rpc_address"`
	} `json:"chain"`
	Handshake struct {
		Enable bool   `json:"enable"`
		Peers  uint64 `json:"peers"`
	}
	Keyring struct {
		Backend string `json:"backend"`
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
		Type             uint64 `json:"type"`
	} `json:"node"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) WithDefaultValues() *Config {
	c.Chain.Gas = 1e5
	c.Chain.GasAdjustment = 0
	c.Chain.GasPrices = "0.1tsent"
	c.Chain.ID = "sentinel-turing-4"
	c.Chain.RPCAddress = "https://rpc.turing.sentinel.co:443"

	c.Handshake.Enable = true
	c.Handshake.Peers = 8

	c.Keyring.Backend = "file"

	c.Node.From = ""
	c.Node.IntervalSessions = 8 * time.Minute.Nanoseconds()
	c.Node.IntervalStatus = 4 * time.Minute.Nanoseconds()
	c.Node.ListenOn = "0.0.0.0:8585"
	c.Node.Moniker = ""
	c.Node.Price = "50tsent"
	c.Node.Provider = ""
	c.Node.RemoteURL = ""
	c.Node.Type = wgtypes.Type

	return c
}

func (c *Config) LoadFromPath(path string) error {
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
	if c.Chain.GasAdjustment < 0 {
		return fmt.Errorf("invalid chain->gas_adjustment; expected non-negative value")
	}
	if c.Chain.ID == "" {
		return fmt.Errorf("invalid chain->id; expected non-empty value")
	}
	if c.Chain.RPCAddress == "" {
		return fmt.Errorf("invalid chain->rpc_address; expected non-empty value")
	}

	if c.Handshake.Peers == 0 {
		return fmt.Errorf("invalid handshake->peers; expected positive value")
	}

	if c.Keyring.Backend == "" {
		return fmt.Errorf("invalid keyring->backend; expected non-empty value")
	}

	if c.Node.From == "" {
		return fmt.Errorf("invalid node->from; expected non-empty value")
	}
	if c.Node.IntervalSessions <= 0 {
		return fmt.Errorf("invalid node->interval_sessions; expected positive value")
	}
	if c.Node.IntervalStatus <= 0 {
		return fmt.Errorf("invalid node->interval_status; expected positive value")
	}
	if c.Node.ListenOn == "" {
		return fmt.Errorf("invalid node->listen_on; expected non-empty value")
	}
	if (c.Node.Provider != "" && c.Node.Price != "") ||
		(c.Node.Provider == "" && c.Node.Price == "") {
		return fmt.Errorf("invalid combination of node->provider and node->price; expected one of them to be empty")
	}
	if c.Node.RemoteURL == "" {
		return fmt.Errorf("invalid node->remote_url; expected non-empty value")
	}
	if c.Node.Type == 0 {
		return fmt.Errorf("invalid node->type; expected positive value")
	}

	return nil
}
