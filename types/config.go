package types

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	hubtypes "github.com/sentinel-official/hub/types"
	"github.com/spf13/viper"

	randutil "github.com/sentinel-official/dvpn-node/utils/rand"
)

const (
	MinPeers                  = 1
	MaxPeers                  = 250
	MinMonikerLength          = 4
	MaxMonikerLength          = 32
	MinIntervalSetSessions    = 10 * time.Second
	MaxIntervalSetSessions    = 10 * time.Minute
	MinIntervalUpdateSessions = (2 * time.Hour) / 2
	MaxIntervalUpdateSessions = (2 * time.Hour) - (5 * time.Minute)
	MinIntervalUpdateStatus   = (1 * time.Hour) / 2
	MaxIntervalUpdateStatus   = (1 * time.Hour) - (5 * time.Minute)
)

var (
	ct = strings.TrimSpace(`
[chain]
# Gas adjustment factor
gas_adjustment = {{ .Chain.GasAdjustment }}

# Gas limit to set per transaction
gas = {{ .Chain.Gas }}

# Gas prices to determine the transaction fee
gas_prices = "{{ .Chain.GasPrices }}"

# The network chain ID
id = "{{ .Chain.ID }}"

# Tendermint RPC interface for the chain
rpc_address = "{{ .Chain.RPCAddress }}"

# Calculate the transaction fee by simulating it
simulate_and_execute = {{ .Chain.SimulateAndExecute }}

[handshake]
# Enable Handshake DNS resolver
enable = {{ .Handshake.Enable }}

# Number of peers
peers = {{ .Handshake.Peers }}

[keyring]
# Underlying storage mechanism for keys
backend = "{{ .Keyring.Backend }}"

# Name of the key with which to sign
from = "{{ .Keyring.From }}"

[node]
# Time interval between each set_sessions operation
interval_set_sessions = "{{ .Node.IntervalSetSessions }}"

# Time interval between each update_sessions transaction
interval_update_sessions = "{{ .Node.IntervalUpdateSessions }}"

# Time interval between each set_status transaction
interval_update_status = "{{ .Node.IntervalUpdateStatus }}"

# API listen-address
listen_on = "{{ .Node.ListenOn }}"

# Name of the node
moniker = "{{ .Node.Moniker }}"

# Per Gigabyte price to charge against the provided bandwidth
price = "{{ .Node.Price }}"

# Address of the provider the node wants to operate under
provider = "{{ .Node.Provider }}"

# Public URL of the node
remote_url = "{{ .Node.RemoteURL }}"

# Type of node
type = "{{ .Node.Type }}"

[qos]
# Limit max number of concurrent peers
max_peers = {{ .QOS.MaxPeers }}
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
	GasAdjustment      float64 `json:"gas_adjustment" mapstructure:"gas_adjustment"`
	GasPrices          string  `json:"gas_prices" mapstructure:"gas_prices"`
	Gas                uint64  `json:"gas" mapstructure:"gas"`
	ID                 string  `json:"id" mapstructure:"id"`
	RPCAddress         string  `json:"rpc_address" mapstructure:"rpc_address"`
	SimulateAndExecute bool    `json:"simulate_and_execute" mapstructure:"simulate_and_execute"`
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

	rpcAddress, err := url.ParseRequestURI(c.RPCAddress)
	if err != nil {
		return errors.Wrap(err, "invalid rpc_address")
	}
	if rpcAddress.Scheme != "http" && rpcAddress.Scheme != "https" {
		return errors.New("rpc_address scheme must be either http or https")
	}
	if rpcAddress.Port() == "" {
		return errors.New("rpc_address port cannot be empty")
	}

	return nil
}

func (c *ChainConfig) WithDefaultValues() *ChainConfig {
	c.GasAdjustment = 1.05
	c.GasPrices = "0.1udvpn"
	c.Gas = 200000
	c.ID = "sentinelhub-2"
	c.RPCAddress = "https://rpc.sentinel.co:443"
	c.SimulateAndExecute = true

	return c
}

type HandshakeConfig struct {
	Enable bool   `json:"enable" mapstructure:"enable"`
	Peers  uint64 `json:"peers" mapstructure:"peers"`
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
	Backend string `json:"backend" mapstructure:"backend"`
	From    string `json:"from" mapstructure:"from"`
}

func NewKeyringConfig() *KeyringConfig {
	return &KeyringConfig{}
}

func (c *KeyringConfig) Validate() error {
	if c.Backend == "" {
		return errors.New("backend cannot be empty")
	}
	if c.Backend != keyring.BackendFile && c.Backend != keyring.BackendTest {
		return fmt.Errorf("backend must be either %s or %s", keyring.BackendFile, keyring.BackendTest)
	}
	if c.From == "" {
		return errors.New("from cannot be empty")
	}

	return nil
}

func (c *KeyringConfig) WithDefaultValues() *KeyringConfig {
	c.Backend = keyring.BackendFile
	c.From = "operator"

	return c
}

type NodeConfig struct {
	IntervalSetSessions    time.Duration `json:"interval_set_sessions" mapstructure:"interval_set_sessions"`
	IntervalUpdateSessions time.Duration `json:"interval_update_sessions" mapstructure:"interval_update_sessions"`
	IntervalUpdateStatus   time.Duration `json:"interval_update_status" mapstructure:"interval_update_status"`
	ListenOn               string        `json:"listen_on" mapstructure:"listen_on"`
	Moniker                string        `json:"moniker" mapstructure:"moniker"`
	Price                  string        `json:"price" mapstructure:"price"`
	Provider               string        `json:"provider" mapstructure:"provider"`
	RemoteURL              string        `json:"remote_url" mapstructure:"remote_url"`
	Type                   string        `json:"type" mapstructure:"type"`
}

func NewNodeConfig() *NodeConfig {
	return &NodeConfig{}
}

func (c *NodeConfig) Validate() error {
	if c.IntervalSetSessions < MinIntervalSetSessions {
		return fmt.Errorf("interval_set_sessions cannot be less than %s", MinIntervalSetSessions)
	}
	if c.IntervalSetSessions > MaxIntervalSetSessions {
		return fmt.Errorf("interval_set_sessions cannot be greater than %s", MaxIntervalSetSessions)
	}
	if c.IntervalUpdateSessions < MinIntervalUpdateSessions {
		return fmt.Errorf("interval_update_sessions cannot be less than %s", MinIntervalUpdateSessions)
	}
	if c.IntervalUpdateSessions > MaxIntervalUpdateSessions {
		return fmt.Errorf("interval_update_sessions cannot be greater than %s", MaxIntervalUpdateSessions)
	}
	if c.IntervalUpdateStatus < MinIntervalUpdateStatus {
		return fmt.Errorf("interval_set_sessions cannot be less than %s", MinIntervalUpdateStatus)
	}
	if c.IntervalUpdateStatus > MaxIntervalUpdateStatus {
		return fmt.Errorf("interval_set_sessions cannot be greater than %s", MaxIntervalUpdateStatus)
	}
	if c.ListenOn == "" {
		return errors.New("listen_on cannot be empty")
	}
	if c.Moniker == "" {
		return errors.New("moniker cannot be empty")
	}
	if len(c.Moniker) < MinMonikerLength {
		return fmt.Errorf("moniker length cannot be less than %d", MinMonikerLength)
	}
	if len(c.Moniker) > MaxMonikerLength {
		return fmt.Errorf("moniker length cannot be greater than %d", MaxMonikerLength)
	}
	if c.Price == "" && c.Provider == "" {
		return errors.New("both price and provider cannot be empty")
	}
	if c.Price != "" && c.Provider != "" {
		return errors.New("either price or provider must be empty")
	}
	if c.Price != "" {
		if _, err := sdk.ParseCoinsNormalized(c.Price); err != nil {
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

	remoteURL, err := url.ParseRequestURI(c.RemoteURL)
	if err != nil {
		return errors.Wrap(err, "invalid remote_url")
	}
	if remoteURL.Scheme != "https" {
		return errors.New("remote_url scheme must be https")
	}
	if remoteURL.Port() == "" {
		return errors.New("remote_url port cannot be empty")
	}

	if c.Type == "" {
		return errors.New("type cannot be empty")
	}
	if c.Type != "wireguard" && c.Type != "v2ray" {
		return errors.New("type must be either wireguard or v2ray")
	}

	return nil
}

func (c *NodeConfig) WithDefaultValues() *NodeConfig {
	c.IntervalSetSessions = 2 * time.Minute
	c.IntervalUpdateSessions = MaxIntervalUpdateSessions
	c.IntervalUpdateStatus = MaxIntervalUpdateStatus
	c.ListenOn = fmt.Sprintf("0.0.0.0:%d", randutil.RandomPort())
	c.Type = "wireguard"

	return c
}

type QOSConfig struct {
	MaxPeers int `json:"max_peers" mapstructure:"max_peers"`
}

func NewQOSConfig() *QOSConfig {
	return &QOSConfig{}
}

func (c *QOSConfig) Validate() error {
	if c.MaxPeers < MinPeers {
		return fmt.Errorf("max_peers cannot be less than %d", MinPeers)
	}
	if c.MaxPeers > MaxPeers {
		return fmt.Errorf("max_peers cannot be greater than %d", MaxPeers)
	}

	return nil
}

func (c *QOSConfig) WithDefaultValues() *QOSConfig {
	c.MaxPeers = MaxPeers

	return c
}

type Config struct {
	Chain     *ChainConfig     `json:"chain" mapstructure:"chain"`
	Handshake *HandshakeConfig `json:"handshake" mapstructure:"handshake"`
	Keyring   *KeyringConfig   `json:"keyring" mapstructure:"keyring"`
	Node      *NodeConfig      `json:"node" mapstructure:"node"`
	QOS       *QOSConfig       `json:"qos" mapstructure:"qos"`
}

func NewConfig() *Config {
	return &Config{
		Chain:     NewChainConfig(),
		Handshake: NewHandshakeConfig(),
		Keyring:   NewKeyringConfig(),
		Node:      NewNodeConfig(),
		QOS:       NewQOSConfig(),
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
	if err := c.QOS.Validate(); err != nil {
		return errors.Wrapf(err, "invalid section qos")
	}

	if c.Node.Type == "v2ray" {
		if c.Handshake.Enable {
			return errors.Wrapf(errors.New("must be disabled"), "invalid section handshake")
		}
	}

	return nil
}

func (c *Config) WithDefaultValues() *Config {
	c.Chain = c.Chain.WithDefaultValues()
	c.Handshake = c.Handshake.WithDefaultValues()
	c.Keyring = c.Keyring.WithDefaultValues()
	c.Node = c.Node.WithDefaultValues()
	c.QOS = c.QOS.WithDefaultValues()

	return c
}

func (c *Config) SaveToPath(path string) error {
	var buffer bytes.Buffer
	if err := t.Execute(&buffer, c); err != nil {
		return err
	}

	return os.WriteFile(path, buffer.Bytes(), 0600)
}

func (c *Config) String() string {
	var buf bytes.Buffer
	if err := t.Execute(&buf, c); err != nil {
		panic(err)
	}

	return buf.String()
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
