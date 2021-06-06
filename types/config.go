package types

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	hubtypes "github.com/sentinel-official/hub/types"
	"github.com/spf13/viper"
)

var (
	ct = strings.TrimSpace(`
[chain]
gas_adjustment = {{ .Chain.GasAdjustment }}
gas = {{ .Chain.Gas }}
gas_prices = "{{ .Chain.GasPrices }}"
id = "{{ .Chain.ID }}"
rpc_address = "{{ .Chain.RPCAddress }}"
simulate_and_execute = {{ .Chain.SimulateAndExecute }}

[handshake]
enable = {{ .Handshake.Enable }}
peers = {{ .Handshake.Peers }}

[keyring]
backend = "{{ .Keyring.Backend }}"
from = "{{ .Keyring.From }}"

[node]
interval_sessions = "{{ .Node.IntervalSessions }}"
interval_status = "{{ .Node.IntervalStatus }}"
listen_on = "{{ .Node.ListenOn }}"
moniker = "{{ .Node.Moniker }}"
price = "{{ .Node.Price }}"
provider = "{{ .Node.Provider }}"
remote_url = "{{ .Node.RemoteURL }}"
	`)

	t = func() *template.Template {
		t, err := template.New("").Parse(ct)
		if err != nil {
			panic(err)
		}

		return t
	}()
)

type ChainConfig struct {
	GasAdjustment      float64 `mapstructure:"gas_adjustment"`
	GasPrices          string  `mapstructure:"gas_prices"`
	Gas                uint64  `mapstructure:"gas"`
	ID                 string  `mapstructure:"id"`
	RPCAddress         string  `mapstructure:"rpc_address"`
	SimulateAndExecute bool    `mapstructure:"simulate_and_execute"`
}

func NewChainConfig() *ChainConfig {
	return &ChainConfig{}
}

func (c *ChainConfig) Validate() error {
	if c.GasAdjustment <= 0 {
		return errors.New("gas_adjustment must be positive")
	}
	if _, err := sdk.ParseCoinsNormalized(c.GasPrices); err != nil {
		return errors.Wrap(err, "invalid gas_prices")
	}
	if c.Gas <= 0 {
		return errors.New("gas must be positive")
	}
	if c.ID == "" {
		return errors.New("id cannot be empty")
	}
	if c.RPCAddress == "" {
		return errors.New("rpc_address cannot be empty")
	}

	return nil
}

func (c *ChainConfig) WithDefaultValues() *ChainConfig {
	c.GasAdjustment = 1.05
	c.GasPrices = "0.1udvpn"
	c.Gas = 200000
	c.ID = ""
	c.RPCAddress = "https://rpc.sentinel.co:443"
	c.SimulateAndExecute = true

	return c
}

type HandshakeConfig struct {
	Enable bool   `mapstructure:"enable"`
	Peers  uint64 `mapstructure:"peers"`
}

func NewHandshakeConfig() *HandshakeConfig {
	return &HandshakeConfig{}
}

func (c *HandshakeConfig) Validate() error {
	if c.Enable {
		if c.Peers <= 0 {
			return errors.New("peers must be positive")
		}
	}

	return nil
}

func (c *HandshakeConfig) WithDefaultValues() *HandshakeConfig {
	c.Enable = true
	c.Peers = 8

	return c
}

type KeyringConfig struct {
	Backend string `mapstructure:"backend"`
	From    string `mapstructure:"from"`
}

func NewKeyringConfig() *KeyringConfig {
	return &KeyringConfig{}
}

func (c *KeyringConfig) Validate() error {
	if c.Backend == "" {
		return errors.New("backend cannot be empty")
	}
	if c.Backend != keyring.BackendFile && c.Backend != keyring.BackendTest {
		return fmt.Errorf("unknown backend %s", c.Backend)
	}
	if c.From == "" {
		return errors.New("from cannot be empty")
	}

	return nil
}

func (c *KeyringConfig) WithDefaultValues() *KeyringConfig {
	c.Backend = keyring.BackendFile

	return c
}

type NodeConfig struct {
	IntervalSessions time.Duration `mapstructure:"interval_sessions"`
	IntervalStatus   time.Duration `mapstructure:"interval_status"`
	ListenOn         string        `mapstructure:"listen_on"`
	Moniker          string        `mapstructure:"moniker"`
	Price            string        `mapstructure:"price"`
	Provider         string        `mapstructure:"provider"`
	RemoteURL        string        `mapstructure:"remote_url"`
}

func NewNodeConfig() *NodeConfig {
	return &NodeConfig{}
}

func (c *NodeConfig) Validate() error {
	if c.IntervalSessions <= 0 {
		return errors.New("interval_sessions must be positive")
	}
	if c.IntervalStatus <= 0 {
		return errors.New("interval_status must be positive")
	}
	if c.ListenOn == "" {
		return errors.New("listen_on cannot be empty")
	}
	if c.Price == "" && c.Provider == "" {
		return errors.New("both price and provider cannot be empty")
	}
	if c.Price != "" && c.Provider != "" {
		return errors.New("either price or provider must be empty")
	}
	if c.Price != "" {
		if _, err := sdk.ParseCoinNormalized(c.Price); err != nil {
			return errors.Wrap(err, "invalid price")
		}
	}
	if c.Provider != "" {
		if _, err := hubtypes.ProvAddressFromBech32(c.Provider); err != nil {
			return errors.Wrap(err, "invalid provider")
		}
	}
	if c.RemoteURL == "" {
		return errors.New("remote_url cannot be empty")
	}

	return nil
}

func (c *NodeConfig) WithDefaultValues() *NodeConfig {
	c.IntervalSessions = 0.9 * 120 * time.Minute
	c.IntervalStatus = 0.9 * 60 * time.Minute
	c.ListenOn = "0.0.0.0:8585"

	return c
}

type Config struct {
	Chain     *ChainConfig     `mapstructure:"chain"`
	Handshake *HandshakeConfig `mapstructure:"handshake"`
	Keyring   *KeyringConfig   `mapstructure:"keyring"`
	Node      *NodeConfig      `mapstructure:"node"`
}

func NewConfig() *Config {
	return &Config{
		Chain:     NewChainConfig(),
		Handshake: NewHandshakeConfig(),
		Keyring:   NewKeyringConfig(),
		Node:      NewNodeConfig(),
	}
}

func (c *Config) Validate() error {
	if err := c.Chain.Validate(); err != nil {
		return errors.Wrapf(err, "invalid section chain")
	}
	if err := c.Handshake.Validate(); err != nil {
		return errors.Wrapf(err, "invalid section handshake")
	}
	if err := c.Keyring.Validate(); err != nil {
		return errors.Wrapf(err, "invalid section keyring")
	}
	if err := c.Node.Validate(); err != nil {
		return errors.Wrapf(err, "invalid section node")
	}

	return nil
}

func (c *Config) WithDefaultValues() *Config {
	c.Chain = c.Chain.WithDefaultValues()
	c.Handshake = c.Handshake.WithDefaultValues()
	c.Keyring = c.Keyring.WithDefaultValues()
	c.Node = c.Node.WithDefaultValues()

	return c
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

func ReadInConfig(v *viper.Viper) (*Config, error) {
	cfg := NewConfig().WithDefaultValues()
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
