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

const (
	managementPort       = 1195
	serverConfigFilePath = "/etc/openvpn/server.conf"
	statusLogFilePath    = "/etc/openvpn/status.log"
)

type OpenVPN struct {
	port             uint16
	publicIP         string
	protocol         string
	encryptionMethod string
	process          *os.Process
}

func NewOpenVPN(port uint16, publicIP, protocol, encryptionMethod string) *OpenVPN {
	return &OpenVPN{
		port:             port,
		publicIP:         publicIP,
		protocol:         protocol,
		encryptionMethod: encryptionMethod,
	}
}

func (o OpenVPN) Type() string {
	return "OpenVPN"
}

func (o OpenVPN) EncryptionMethod() string {
	return o.encryptionMethod
}

func (o OpenVPN) writeServerConfig() error {
	data := fmt.Sprintf(serverConfigTemplate, o.port, o.encryptionMethod, statusLogFilePath, managementPort)

	return ioutil.WriteFile(serverConfigFilePath, []byte(data), os.ModePerm)
}

func (o OpenVPN) generateServerKeys() error {
	cmd := exec.Command("sh", "-c", commandGenerateServerKeys)

	log.Printf("Generating the OpenVPN server keys, it will take some time...")
	return cmd.Run()
}

func (o OpenVPN) setNATRouting() error {
	cmd := exec.Command("sh", "-c", commandNATRouting)

	return cmd.Run()
}

func (o OpenVPN) Init() error {
	log.Printf("Initializing the OpenVPN server")
	if err := o.generateServerKeys(); err != nil {
		return err
	}

	if err := o.writeServerConfig(); err != nil {
		return err
	}

	return o.setNATRouting()
}

func (o OpenVPN) Start() error {
	cmd := exec.Command("openvpn", "--config", serverConfigFilePath)

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
		return errors.Errorf("OpenVPN process is nil")
	}

	return o.process.Kill()
}

func (o OpenVPN) ClientList() (map[string]sdkTypes.Bandwidth, error) {
	if _, err := os.Stat(statusLogFilePath); os.IsNotExist(err) {
		log.Printf("OpenVPN status log file does not exist")
		return nil, nil
	}

	file, err := os.Open(statusLogFilePath)
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

func (o OpenVPN) RevokeClientCertificate(cname string) error {
	cmd := exec.Command("sh", "-c", commandRevokeClientCertificate(cname))

	log.Printf("Revoking the client certificate with common name `%s`", cname)
	return cmd.Run()
}

func (o OpenVPN) DisconnectClient(id string) error {
	cname := "client_" + id
	cmd := exec.Command("sh", "-c", commandDisconnectClient(cname, managementPort))

	log.Printf("Disconnecting the client with common name `%s`", cname)
	if err := cmd.Run(); err != nil {
		return err
	}

	return o.RevokeClientCertificate(cname)
}

func (o OpenVPN) GenerateClientKey(id string) ([]byte, error) {
	cname := "client_" + id
	certPath := fmt.Sprintf("/usr/share/easy-rsa/pki/issued/%s.crt", cname)
	keyPath := fmt.Sprintf("/usr/share/easy-rsa/pki/private/%s.key", cname)
	_, certPathErr := os.Stat(certPath)
	_, keyPathErr := os.Stat(keyPath)

	if os.IsNotExist(certPathErr) || os.IsNotExist(keyPathErr) {
		cmd := exec.Command("sh", "-c", commandGenerateClientKeys(cname))

		log.Printf("Generating the client key with common name `%s`", cname)
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

	clientKey := fmt.Sprintf(clientConfigTemplate, o.publicIP, o.port, o.encryptionMethod, ca, cert, key, tlsAuth)

	return []byte(clientKey), nil
}
