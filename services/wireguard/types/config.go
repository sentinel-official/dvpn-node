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
# Name of the network interface
interface = "{{ .Interface }}"

# Port number to accept the incoming connections
listen_port = {{ .ListenPort }}

# Server private key
private_key = "{{ .PrivateKey }}"
	`)

	t = func() *template.Template {
		t, err := template.New("wireguard_toml").Parse(ct)
		if err != nil {
			panic(err)
		}

		return t
	}()
)

type Config struct {
	Interface  string `json:"interface" mapstructure:"interface"`
	ListenPort uint16 `json:"listen_port" mapstructure:"listen_port"`
	PrivateKey string `json:"private_key" mapstructure:"private_key"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Validate() error {
	if c.Interface == "" {
		return fmt.Errorf("interface cannot be empty")
	}
	if c.ListenPort == 0 {
		return fmt.Errorf("listen_port cannot be zero")
	}
	if c.PrivateKey == "" {
		return fmt.Errorf("private_key cannot be empty")
	}
	if _, err := KeyFromString(c.PrivateKey); err != nil {
		return errors.Join(err, fmt.Errorf("invalid private_key"))
	}

	return nil
}

func (c *Config) WithDefaultValues() *Config {
	key, err := NewPrivateKey()
	if err != nil {
		panic(err)
	}

	c.Interface = "wg0"
	c.ListenPort = randutil.RandomPort()
	c.PrivateKey = key.String()

	return c
}

func (c *Config) SaveToPath(path string) error {
	var buffer bytes.Buffer
	if err := t.Execute(&buffer, c); err != nil {
		return err
	}

	return os.WriteFile(path, buffer.Bytes(), 0644)
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
