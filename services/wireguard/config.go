package wireguard

import (
	"bytes"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"text/template"

	"github.com/pelletier/go-toml"
)

var (
	ct = strings.TrimSpace(`
device = "{{ .Device }}"
listen_port = {{ .ListenPort }}
private_key = "{{ .PrivateKey }}"
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
	Device     string `json:"device"`
	ListenPort uint16 `json:"listen_port"`
	PrivateKey string `json:"private_key"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) WithDefaultValues() *Config {
	c.Device = "wg0"
	c.ListenPort = uint16(rand.Int31n(1<<16-1<<10) + 1<<10)

	key := make([]byte, 32)
	if _, err := crand.Read(key); err != nil {
		panic(err)
	}

	c.PrivateKey = base64.StdEncoding.EncodeToString(key)

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
	if c.Device == "" {
		return fmt.Errorf("device is empty")
	}
	if c.ListenPort == 0 {
		return fmt.Errorf("listen_port is zero")
	}
	if c.PrivateKey == "" {
		return fmt.Errorf("private_key is empty")
	}

	return nil
}
