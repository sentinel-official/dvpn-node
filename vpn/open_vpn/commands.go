package open_vpn

import (
	"fmt"
)

func cmdGenerateClientKeys(cname string) string {
	return fmt.Sprintf(generateClientKeysCommandTemplate, cname)
}

func cmdDisconnectClient(cname string, managementPort uint32) string {
	return fmt.Sprintf("echo 'kill %s' | nc 127.0.0.1 %d", cname, managementPort)
}

func cmdRevokeClientCert(cname string) string {
	return fmt.Sprintf(revokeClientCertCommandTemplate, cname)
}

var cmdGenerateServerKeys = `
cd /usr/share/easy-rsa && \
./easyrsa init-pki && \
echo \r | ./easyrsa build-ca nopass && \
./easyrsa gen-dh && \
./easyrsa gen-crl && \
./easyrsa build-server-full server nopass && \
openvpn --genkey --secret pki/ta.key && \
chmod 755 pki/crl.pem
`
