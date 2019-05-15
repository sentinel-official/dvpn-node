package open_vpn

var serverConfigTemplate = `
dev tun
port %d
proto udp

ca /usr/share/easy-rsa/pki/ca.crt
cert /usr/share/easy-rsa/pki/issued/server.crt
key /usr/share/easy-rsa/pki/private/server.key
dh /usr/share/easy-rsa/pki/dh.pem
server 10.8.0.0 255.255.255.0

push "redirect-gateway def1 bypass-dhcp"
push "dhcp-option DNS 208.67.222.222"
push "dhcp-option DNS 208.67.220.220"
keepalive 10 120
tls-auth /usr/share/easy-rsa/pki/ta.key 0
cipher %s
persist-key
persist-tun
status %s 2
verb 3

auth SHA256
crl-verify /usr/share/easy-rsa/pki/crl.pem
key-direction 0
management 127.0.0.1 %d
`

var clientConfigTemplate = `
client

dev tun
proto udp
remote %s %d
resolv-retry infinite
nobind
persist-key
persist-tun
remote-cert-tls server
cipher %s
verb 3
auth SHA256
key-direction 1
script-security 2

<ca>
%s
</ca>
<cert>
%s
</cert>
<key>
%s
</key>
<tls-auth>
%s
</tls-auth>
`
