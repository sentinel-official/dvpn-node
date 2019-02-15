package open_vpn

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/ironman0x7b2/vpn-node/config"
	"github.com/ironman0x7b2/vpn-node/vpn"
)

type OpenVPN struct {
	ipAddress   string
	port        uint32
	protocol    string
	encryption  string
	process     *os.Process
	processWait chan error
}

func NewOpenVPNFromAppConfig(appCfg *config.AppConfig) OpenVPN {
	return OpenVPN{
		ipAddress:   appCfg.VPN.IPAddress,
		port:        appCfg.VPN.Port,
		protocol:    appCfg.VPN.Protocol,
		encryption:  appCfg.VPN.EncryptionMethod,
		processWait: make(chan error, 1),
	}
}

func (o OpenVPN) Type() string {
	return "OpenVPN"
}

func (o OpenVPN) Encryption() string {
	return o.encryption
}

func (o OpenVPN) WriteServerConfig() error {
	t, err := template.New("server_config").Parse(serverConfigTemplate)
	if err != nil {
		return err
	}

	var stdout bytes.Buffer
	if err := t.Execute(&stdout, newServerConfigData(o.port, o.protocol, o.encryption)); err != nil {
		return err
	}

	return ioutil.WriteFile(defaultServerConfigFilePath, stdout.Bytes(), os.ModePerm)
}

func (o OpenVPN) WriteClientConfig() error {
	t, err := template.New("client_config").Parse(clientConfigTemplate)
	if err != nil {
		return err
	}

	var stdout bytes.Buffer
	if err := t.Execute(&stdout, newClientConfigData(o.ipAddress, o.port, o.protocol, o.encryption)); err != nil {
		return err
	}

	return ioutil.WriteFile(defaultClientConfigFilePath, stdout.Bytes(), os.ModePerm)
}

func (o OpenVPN) Wait(done chan error) {
	err := <-o.processWait
	done <- err
}

func (o OpenVPN) Init() error {
	if err := o.WriteServerConfig(); err != nil {
		return err
	}

	return o.WriteClientConfig()
}

func (o OpenVPN) Start() error {
	cmd := exec.Command("openvpn", "--config", defaultServerConfigFilePath)
	if err := cmd.Start(); err != nil {
		return err
	}

	o.process = cmd.Process
	go func() {
		o.processWait <- cmd.Wait()
	}()

	return nil
}

func (o OpenVPN) Stop() error {
	if o.process == nil {
		return fmt.Errorf("process is nil")
	}

	return o.process.Kill()
}

func (o OpenVPN) ClientList() ([]vpn.Client, error) {
	if _, err := os.Stat(defaultStatusLogFilePath); os.IsNotExist(err) {
		return nil, nil
	}

	file, err := os.Open(defaultStatusLogFilePath)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	var clients []vpn.Client
	reader := bufio.NewReader(file)

	for {
		lineBytes, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		line := string(lineBytes)
		if strings.Contains(line, cnamePrefix) {
			lineSlice := strings.Split(line, ",")
			cname := lineSlice[0][len(cnamePrefix):]

			upload, err := strconv.Atoi(lineSlice[2])
			if err != nil {
				return nil, err
			}

			download, err := strconv.Atoi(lineSlice[3])
			if err != nil {
				return nil, err
			}

			client := vpn.NewClient(cname, upload, download)
			clients = append(clients, client)
		} else if strings.Contains(line, "ROUTING TABLE") {
			break
		}
	}

	return clients, nil
}

func (o OpenVPN) RevokeClientCert(cname string) error {
	cmdParams, err := cmdRevokeClientCert(cname)
	if err != nil {
		return err
	}

	cmd := exec.Command("sh", "-c", cmdParams)

	return cmd.Run()
}

func (o OpenVPN) DisconnectClient(id string) error {
	cname := cnamePrefix + id
	cmd := exec.Command("sh", "-c",
		fmt.Sprintf("echo 'kill %s' | nc 127.0.0.1 1195", cname))

	if err := cmd.Run(); err != nil {
		return err
	}

	return o.RevokeClientCert(cname)
}

func (o OpenVPN) GenerateClientKey(id string) ([]byte, error) {
	cname := cnamePrefix + id
	ovpnFilePath := filepath.Join(defaultKeysDir, cname+ovpnFileExtension)

	cmdParams, err := cmdGenerateClientKey(cname)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("sh", "-c", cmdParams)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return o.ReadOVPNFile(ovpnFilePath)
}

func (o OpenVPN) ReadOVPNFile(filePath string) ([]byte, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	return ioutil.ReadAll(file)
}
