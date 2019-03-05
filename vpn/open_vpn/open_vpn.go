package open_vpn

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
	"github.com/pkg/errors"
)

type OpenVPN struct {
	port           uint16
	managementPort uint16
	ip             string
	protocol       string
	encryption     string
	process        *os.Process
}

func NewOpenVPN(port, managementPort uint16, ip, protocol, encryption string) *OpenVPN {
	return &OpenVPN{
		port:           port,
		managementPort: managementPort,
		ip:             ip,
		protocol:       protocol,
		encryption:     encryption,
	}
}

func (o OpenVPN) Type() string {
	return "OpenVPN"
}

func (o OpenVPN) Encryption() string {
	return o.encryption
}

func (o OpenVPN) writeServerConfig() error {
	data := fmt.Sprintf(serverConfigTemplate, o.port, o.encryption, o.managementPort)

	log.Println("Writing OpenVPN server config to path /etc/openvpn/server.conf")
	return ioutil.WriteFile("/etc/openvpn/server.conf", []byte(data), os.ModePerm)
}

func (o OpenVPN) generateServerKeys() error {
	cmd := exec.Command("sh", "-c", cmdGenerateServerKeys)

	log.Println("Generating OpenVPN server keys, it will take some time")
	return cmd.Run()
}

func (o OpenVPN) setIPTablesRules() error {
	cmd := exec.Command("sh", "-c", cmdIPTables)

	log.Println("Setting up OpenVPN IPTables rules")
	return cmd.Run()
}

func (o OpenVPN) Init() error {
	log.Println("Initializing the OpenVPN")
	if err := o.setIPTablesRules(); err != nil {
		return err
	}

	if err := o.generateServerKeys(); err != nil {
		return err
	}

	return o.writeServerConfig()
}

func (o OpenVPN) Start() error {
	cmd := exec.Command("openvpn", "--config", "/etc/openvpn/server.conf")

	log.Printf("Starting OpenVPN server with the config file /etc/openvpn/server.conf")
	if err := cmd.Start(); err != nil {
		return err
	}

	o.process = cmd.Process
	go func() {
		if err := cmd.Wait(); err != nil {
			panic(err)
		}
	}()

	return nil
}

func (o OpenVPN) Stop() error {
	if o.process == nil {
		return errors.New("Process is nil")
	}

	return o.process.Kill()
}

func (o OpenVPN) ClientList() (map[string]sdkTypes.Bandwidth, error) {
	filePath := "/etc/openvpn/openvpn-status.log"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Println("OpenVPN status log file not exist")
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

	clients := make(map[string]sdkTypes.Bandwidth)
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
			id := lineSlice[0][len("client_"):]

			upload, err := strconv.Atoi(lineSlice[2])
			if err != nil {
				return nil, err
			}

			download, err := strconv.Atoi(lineSlice[3])
			if err != nil {
				return nil, err
			}

			clients[id] = sdkTypes.NewBandwidthFromInt64(int64(upload), int64(download))
		} else if strings.Contains(line, "ROUTING TABLE") {
			break
		}
	}

	return clients, nil
}

func (o OpenVPN) RevokeClientCert(cname string) error {
	cmd := exec.Command("sh", "-c", cmdRevokeClientCert(cname))

	log.Printf("Revoking client certificate with common name `%s`", cname)
	return cmd.Run()
}

func (o OpenVPN) DisconnectClient(id string) error {
	cname := "client_" + id
	cmd := exec.Command("sh", "-c", cmdDisconnectClient(cname, o.managementPort))

	log.Printf("Disconnecting client with common name `%s`", cname)
	if err := cmd.Run(); err != nil {
		return err
	}

	return o.RevokeClientCert(cname)
}

func (o OpenVPN) GenerateClientKey(id string) ([]byte, error) {
	cname := "client_" + id
	certPath := fmt.Sprintf("/usr/share/easy-rsa/pki/issued/%s.crt", cname)
	keyPath := fmt.Sprintf("/usr/share/easy-rsa/pki/private/%s.key", cname)
	_, certPathErr := os.Stat(certPath)
	_, keyPathErr := os.Stat(keyPath)

	if os.IsNotExist(certPathErr) || os.IsNotExist(keyPathErr) {
		cmd := exec.Command("sh", "-c", cmdGenerateClientKeys(cname))

		log.Printf("Generating client key with common name `%s`", cname)
		if err := cmd.Run(); err != nil {
			return nil, err
		}
	}

	ca, err := ioutil.ReadFile("/usr/share/easy-rsa/pki/ca.crt")
	if err != nil {
		return nil, err
	}

	cert, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	key, err := ioutil.ReadFile(keyPath)
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
