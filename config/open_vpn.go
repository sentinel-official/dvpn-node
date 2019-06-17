package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	"github.com/pelletier/go-toml"
	"github.com/tendermint/tendermint/libs/common"

	"github.com/ironman0x7b2/vpn-node/types"
)

// nolint: gochecknoglobals
var (
	openVPNConfigTemplate        *template.Template
	defaultOpenVPNConfigTemplate = `
port = {{ .Port }}
protocol = "{{ .Protocol }}"
encryption = "{{ .Encryption }}"
`
)

// nolint: gochecknoinits
func init() {
	var err error

	openVPNConfigTemplate, err = template.New("openVPNConfig").Parse(defaultOpenVPNConfigTemplate)
	if err != nil {
		panic(err)
	}
}

type OpenVPNConfig struct {
	Port       uint16 `json:"port"`
	Protocol   string `json:"protocol"`
	Encryption string `json:"encryption"`
}

func NewOpenVPNConfig() *OpenVPNConfig {
	return &OpenVPNConfig{}
}

// nolint:dupl
func (o *OpenVPNConfig) LoadFromPath(path string) error {
	if path == "" {
		path = types.DefaultOpenVPNConfigFilePath
	}

	log.Printf("Loading the OpenVPN configuration from path `%s`", path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		*o = OpenVPNConfig{}
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

	return json.Unmarshal(data, o)
}

func (o *OpenVPNConfig) SaveToPath(path string) error {
	var buffer bytes.Buffer
	if err := openVPNConfigTemplate.Execute(&buffer, o); err != nil {
		return err
	}

	if path == "" {
		path = types.DefaultOpenVPNConfigFilePath
	}

	return common.WriteFile(path, buffer.Bytes(), os.ModePerm)
}
