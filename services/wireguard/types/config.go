package types

import (
	"bytes"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"text/template"

	"github.com/pelletier/go-toml"
)

var (
	ct = strings.TrimSpace(`
interface = "{{ .Interface }}"
listen_port = {{ .ListenPort }}
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
	Interface  string `json:"interface"`
	ListenPort uint16 `json:"listen_port"`
	PrivateKey string `json:"private_key"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) WithDefaultValues() *Config {
	c.Interface = "wg0"

	n, _ := crand.Int(crand.Reader, big.NewInt(1<<16-1<<10))
	c.ListenPort = uint16(n.Int64() + 1<<10)

	key, err := NewPrivateKey()
	if err != nil {
		panic(err)
	}

	c.PrivateKey = key.String()

	return c
}

func (c *Config) LoadFromPath(path string) error {
	if _, err := os.Stat(path); err != nil {
		config := NewConfig().WithDefaultValues()
		if err := config.SaveToPath(path); err != nil {
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
	if c.Interface == "" {
		return fmt.Errorf("invalid interface")
	}
	if c.ListenPort == 0 {
		return fmt.Errorf("invalid listen_port")
	}
	if c.PrivateKey == "" {
		return fmt.Errorf("invalid private_key")
	}

	return nil
}
