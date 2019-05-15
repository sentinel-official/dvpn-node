package open_vpn

import (
	"fmt"
)

func commandGenerateClientKeys(cname string) string {
	return fmt.Sprintf("cat /dev/null > /usr/share/easy-rsa/pki/index.txt && "+
		"/usr/share/easy-rsa/easyrsa build-client-full %s nopass", cname)
}

func commandDisconnectClient(cname string, managementPort uint16) string {
	return fmt.Sprintf("echo 'kill %s' | nc 127.0.0.1 %d", cname, managementPort)
}

func commandRevokeClientCertificate(cname string) string {
	return fmt.Sprintf("echo yes | /usr/share/easy-rsa/easyrsa revoke %s && "+
		"/usr/share/easy-rsa/easyrsa gen-crl", cname)
}

var commandGenerateServerKeys = `
rm -rf /usr/share/easy-rsa/pki &&
/usr/share/easy-rsa/easyrsa init-pki &&
echo \r | /usr/share/easy-rsa/easyrsa build-ca nopass &&
/usr/share/easy-rsa/easyrsa gen-dh &&
/usr/share/easy-rsa/easyrsa gen-crl &&
/usr/share/easy-rsa/easyrsa build-server-full server nopass &&
openvpn --genkey --secret /usr/share/easy-rsa/pki/ta.key
`

var commandNATRouting = `
iptables -t nat -C POSTROUTING -s 10.8.0.0/8 -o eth0 -j MASQUERADE || {
	iptables -t nat -A POSTROUTING -s 10.8.0.0/8 -o eth0 -j MASQUERADE
}
`
