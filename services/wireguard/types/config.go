package types

import (
	"bytes"
	"io/ioutil"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	ct = strings.TrimSpace(`
# Name of the WireGuard interface
iface = "{{ .IFace }}"

# Name of the WAN interface
iface_wan = "{{ .IFaceWAN }}"

# IPv4 CIDR block for peers
ipv4_cidr = "{{ .IPv4CIDR }}"

# IPv6 CIDR block for peers
ipv6_cidr = "{{ .IPv6CIDR }}"

# Port number to accept the incoming connections
listen_port = {{ .ListenPort }}

# Server private key
private_key = "{{ .PrivateKey }}"
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
	IFace      string `json:"iface" mapstructure:"iface"`
	IFaceWAN   string `json:"iface_wan" mapstructure:"iface_wan"`
	IPv4CIDR   string `json:"ipv4_cidr" mapstructure:"ipv4_cidr"`
	IPv6CIDR   string `json:"ipv6_cidr" mapstructure:"ipv6_cidr"`
	ListenPort uint16 `json:"listen_port" mapstructure:"listen_port"`
	PrivateKey string `json:"private_key" mapstructure:"private_key"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Validate() error {
	if c.IFace == "" {
		return errors.New("iface cannot be empty")
	}
	if c.IFaceWAN == "" {
		return errors.New("iface_wan cannot be empty")
	}
	if c.IPv4CIDR == "" {
		return errors.New("ipv4_cidr cannot be empty")
	}
	if c.IPv6CIDR == "" {
		return errors.New("ipv6_cidr cannot be empty")
	}
	if c.ListenPort == 0 {
		return errors.New("listen_port cannot be zero")
	}
	if c.PrivateKey == "" {
		return errors.New("private_key cannot be empty")
	}
	if _, err := KeyFromString(c.PrivateKey); err != nil {
		return errors.Wrap(err, "invalid private_key")
	}

	return nil
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
