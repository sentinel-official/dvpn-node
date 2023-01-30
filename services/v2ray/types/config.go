package types

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	randutil "github.com/sentinel-official/dvpn-node/utils/rand"
)

var (
	ct = strings.TrimSpace(`
# Name of the transmission protocol
protocol = "{{ .Protocol }}"

[vmess]
# Port number to accept the incoming connections
listen_port = {{ .VMess.ListenPort }}
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
}

func NewVMessConfig() *VMessConfig {
	return &VMessConfig{}
}

func (c *VMessConfig) WithDefaultValues() *VMessConfig {
	c.ListenPort = randutil.RandomPort()

	return c
}

func (c *VMessConfig) Validate() error {
	if c.ListenPort == 0 {
		return errors.New("listen_port cannot be zero")
	}

	return nil
}

type Config struct {
	Protocol string       `json:"protocol" mapstructure:"protocol"`
	VMess    *VMessConfig `json:"vmess" mapstructure:"vmess"`
}

func NewConfig() *Config {
	return &Config{
		VMess: NewVMessConfig(),
	}
}

func (c *Config) Validate() error {
	if c.Protocol == "" {
		return errors.New("protocol cannot be empty")
	}
	if c.Protocol != "vmess" {
		return errors.New("protocol must be vmess")
	}

	if err := c.VMess.Validate(); err != nil {
		return errors.Wrapf(err, "invalid section vmess")
	}

	return nil
}

func (c *Config) WithDefaultValues() *Config {
	c.Protocol = "vmess"

	c.VMess = c.VMess.WithDefaultValues()

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
