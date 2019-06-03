package config

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	"github.com/pelletier/go-toml"
	"github.com/tendermint/tendermint/libs/common"

	"github.com/ironman0x7b2/vpn-node/types"
)

var openVPNConfigTemplate *template.Template

func init() {
	var err error

	openVPNConfigTemplate, err = template.New("openVPNConfig").Parse(defaultOpenVPNConfigTemplate)
	if err != nil {
		panic(err)
	}
}

var defaultOpenVPNConfigTemplate = `
##### base options #####
port = {{ .Port }}
protocol = "{{ .Protocol }}"
encryption = "{{ .Encryption }}"
`

type OpenVPNConfig struct {
	Port       uint16 `json:"port"`
	Protocol   string `json:"protocol"`
	Encryption string `json:"encryption"`
}

func NewOpenVPNConfig() *OpenVPNConfig {
	return &OpenVPNConfig{}
}

func (o *OpenVPNConfig) LoadFromPath(path string) error {
	if len(path) == 0 {
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

	return toml.Unmarshal(data, o)
}

func (o OpenVPNConfig) SaveToPath(path string) error {
	var buffer bytes.Buffer
	if err := openVPNConfigTemplate.Execute(&buffer, o); err != nil {
		return err
	}

	if len(path) == 0 {
		path = types.DefaultOpenVPNConfigFilePath
	}

	return common.WriteFile(path, buffer.Bytes(), os.ModePerm)
}
