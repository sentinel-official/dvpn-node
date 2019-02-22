package open_vpn

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/ironman0x7b2/vpn-node/types"
)

type OpenVPN struct {
	ip             string
	port           uint32
	protocol       string
	encryption     string
	managementPort uint32
	process        *os.Process
	processWait    chan error
}

func NewOpenVPN(ip string, port uint32, protocol, encryption string, managementPort uint32) OpenVPN {
	return OpenVPN{
		ip:             ip,
		port:           port,
		protocol:       protocol,
		encryption:     encryption,
		managementPort: managementPort,
		processWait:    make(chan error),
	}
}

func (o OpenVPN) Type() string {
	return "OpenVPN"
}

func (o OpenVPN) Encryption() string {
	return o.encryption
}

func (o OpenVPN) WriteServerConfig() error {
	data := fmt.Sprintf(serverConfigTemplate, o.port, o.encryption, o.managementPort)

	return ioutil.WriteFile("/etc/openvpn/server.conf", []byte(data), os.ModePerm)
}

func (o OpenVPN) GenerateServerKeys() error {
	cmd := exec.Command("sh", "-c", cmdGenerateServerKeys)

	return cmd.Run()
}

func (o OpenVPN) Wait(done chan error) {
	err := <-o.processWait
	done <- err
}

func (o OpenVPN) Init() error {
	if err := o.GenerateServerKeys(); err != nil {
		return err
	}

	return o.WriteServerConfig()
}

func (o OpenVPN) Start() error {
	cmd := exec.Command("openvpn", "--config", "/etc/openvpn/server.conf")
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
		return errors.New("Process is nil")
	}

	return o.process.Kill()
}

func (o OpenVPN) ClientList() (types.VPNClients, error) {
	filePath := "/etc/openvpn/openvpn-status.log"
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

	var clients types.VPNClients
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
		if strings.Contains(line, "client_") {
			lineSlice := strings.Split(line, ",")
			cname := lineSlice[0][len("client_"):]

			upload, err := strconv.Atoi(lineSlice[2])
			if err != nil {
				return nil, err
			}

			download, err := strconv.Atoi(lineSlice[3])
			if err != nil {
				return nil, err
			}

			client := types.NewVPNClient(cname, int64(upload), int64(download))
			clients = clients.Append(client)
		} else if strings.Contains(line, "ROUTING TABLE") {
			break
		}
	}

	return clients.Sort(), nil
}

func (o OpenVPN) RevokeClientCert(cname string) error {
	cmd := exec.Command("sh", "-c", cmdRevokeClientCert(cname))

	return cmd.Run()
}

func (o OpenVPN) DisconnectClient(id string) error {
	cname := "client_" + id
	cmd := exec.Command("sh", "-c", cmdDisconnectClient(cname, o.managementPort))

	if err := cmd.Run(); err != nil {
		return err
	}

	return o.RevokeClientCert(cname)
}

func (o OpenVPN) GenerateClientKey(id string) ([]byte, error) {
	cname := "client_" + id

	cmd := exec.Command("sh", "-c", cmdGenerateClientKeys(cname))
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	ca, err := ioutil.ReadFile("/usr/share/easy-rsa/pki/ca.crt")
	if err != nil {
		return nil, err
	}

	cert, err := ioutil.ReadFile(fmt.Sprintf("/usr/share/easy-rsa/pki/issued/%s.crt", cname))
	if err != nil {
		return nil, err
	}

	key, err := ioutil.ReadFile(fmt.Sprintf("/usr/share/easy-rsa/pki/private/%s.key", cname))
	if err != nil {
		return nil, err
	}

	tlsAuth, err := ioutil.ReadFile("/usr/share/easy-rsa/pki/ta.key")
	if err != nil {
		return nil, err
	}

	ovpn := fmt.Sprintf(clientConfigTemplate, o.ip, o.port, o.encryption, ca, cert, key, tlsAuth)

	return []byte(ovpn), nil
}
