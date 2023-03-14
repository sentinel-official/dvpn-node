package types

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/viper"

	randutil "github.com/sentinel-official/dvpn-node/utils/rand"
)

var (
	ct = strings.TrimSpace(`
[vmess]
# Port number to accept the incoming connections
listen_port = {{ .VMess.ListenPort }}

# Name of the transport protocol
transport = "{{ .VMess.Transport }}"
	`)

	t = func() *template.Template {
		t, err := template.New("v2ray_toml").Parse(ct)
		if err != nil {
			panic(err)
		}

		return t
	}()
)

type VMessConfig struct {
	ListenPort uint16 `json:"listen_port" mapstructure:"listen_port"`
	Transport  string `json:"transport" mapstructure:"transport"`
}

func NewVMessConfig() *VMessConfig {
	return &VMessConfig{}
}

func (c *VMessConfig) WithDefaultValues() *VMessConfig {
	c.ListenPort = randutil.RandomPort()
	c.Transport = "grpc"

	return c
}

func (c *VMessConfig) Validate() error {
	if c.ListenPort == 0 {
		return fmt.Errorf("listen_port cannot be zero")
	}
	if c.Transport == "" {
		return fmt.Errorf("transport cannot be empty")
	}

	t := NewTransportFromString(c.Transport)
	if !t.IsValid() {
		return fmt.Errorf("invalid transport")
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
		return errors.Join(err, fmt.Errorf("invalid section vmess"))
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

	return os.WriteFile(path, buf.Bytes(), 0644)
}

func (c *Config) String() string {
	var buf bytes.Buffer
	if err := t.Execute(&buf, c); err != nil {
		panic(err)
	}

	return buf.String()
}

func ReadInConfig(v *viper.Viper) (*Config, error) {
	config := NewConfig().WithDefaultValues()
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := v.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
