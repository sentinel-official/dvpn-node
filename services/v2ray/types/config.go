package types

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	tmstrings "github.com/tendermint/tendermint/libs/strings"

	randutil "github.com/sentinel-official/dvpn-node/utils/rand"
)

var (
	ct = strings.TrimSpace(`
[vmess]
# Port number to accept the incoming connections
listen_port = {{ .VMess.ListenPort }}

# Name of the transport protocol
protocol = "{{ .VMess.Protocol }}"
	`)

	t = func() *template.Template {
		t, err := template.New("config_v2ray_toml").Parse(ct)
		if err != nil {
			panic(err)
		}

		return t
	}()
)

type VMessConfig struct {
	ListenPort uint16 `json:"listen_port" mapstructure:"listen_port"`
	Protocol   string `json:"protocol" mapstructure:"protocol"`
}

func NewVMessConfig() *VMessConfig {
	return &VMessConfig{}
}

func (c *VMessConfig) WithDefaultValues() *VMessConfig {
	c.ListenPort = randutil.RandomPort()
	c.Protocol = "grpc"

	return c
}

func (c *VMessConfig) Validate() error {
	if c.ListenPort == 0 {
		return errors.New("listen_port cannot be zero")
	}
	if c.Protocol == "" {
		return errors.New("protocol cannot be empty")
	}
	if !tmstrings.StringInSlice(c.Protocol, TransportProtocols) {
		return fmt.Errorf("protocol must be one of %#v", TransportProtocols)
	}

	return nil
}

type Config struct {
	VMess *VMessConfig `json:"vmess" mapstructure:"vmess"`
}

func NewConfig() *Config {
	return &Config{
		VMess: NewVMessConfig(),
	}
}

func (c *Config) Validate() error {
	if err := c.VMess.Validate(); err != nil {
		return errors.Wrapf(err, "invalid section vmess")
	}

	return nil
}

func (c *Config) WithDefaultValues() *Config {
	c.VMess = c.VMess.WithDefaultValues()

	return c
}

func (c *Config) SaveToPath(path string) error {
	var buf bytes.Buffer
	if err := t.Execute(&buf, c); err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0600)
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
