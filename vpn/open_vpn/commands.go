// nolint:gochecknoglobals
package open_vpn

import (
	"fmt"
)

func commandGenerateClientKeys(cname string) string {
	return fmt.Sprintf("cd /usr/share/easy-rsa/ && "+
		"cat /dev/null > pki/index.txt && "+
		"./easyrsa build-client-full %s nopass", cname)
}

func commandDisconnectClient(cname string, managementPort uint16) string {
	return fmt.Sprintf("echo 'kill %s' | nc 127.0.0.1 %d", cname, managementPort)
}

func commandRevokeClientCertificate(cname string) string {
	return fmt.Sprintf("cd /usr/share/easy-rsa/ && "+
		"echo yes | ./easyrsa revoke %s && "+
		"./easyrsa gen-crl", cname)
}

var commandGenerateServerKeys = `
cd /usr/share/easy-rsa/ &&
rm -rf pki/ &&
./easyrsa init-pki &&
echo \r | ./easyrsa build-ca nopass &&
./easyrsa gen-dh &&
./easyrsa gen-crl &&
./easyrsa build-server-full server nopass &&
openvpn --genkey --secret pki/ta.key
`

var commandNATRouting = `
iptables -t nat -C POSTROUTING -s 10.8.0.0/8 -o eth0 -j MASQUERADE || {
	iptables -t nat -A POSTROUTING -s 10.8.0.0/8 -o eth0 -j MASQUERADE
}
`
