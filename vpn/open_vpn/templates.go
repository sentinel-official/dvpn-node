package open_vpn

var generateClientKeyCommandTemplate = `
cd {{ .EasyRSADir}} && \
cat /dev/null > {{ .EasyRSADir}}/pki/index.txt && \
./easyrsa build-client-full {{ .CName}} nopass && \

cd pki && \
chmod 755 crl.pem && \
cp *.crt *.key *.pem private/*.key issued/*.crt reqs/*.req {{ .KeysDir}} && \

cat {{ .ClientConfigFilePath}} > {{ .OVPNFilePath}} && \
echo '<ca>' >> {{ .OVPNFilePath}} && \
cat {{ .KeysDir}}/ca.crt >> {{ .OVPNFilePath}} && \
echo '</ca>' >> {{ .OVPNFilePath}} && \
echo '<cert>' >> {{ .OVPNFilePath}} && \
cat {{ .KeysDir}}/{{ .CName}}.crt >> {{ .OVPNFilePath}} && \
echo '</cert>' >> {{ .OVPNFilePath}} && \
echo '<key>' >> {{ .OVPNFilePath}} && \
cat {{ .KeysDir}}/{{ .CName}}.key >> {{ .OVPNFilePath}} && \
echo '</key>' >> {{ .OVPNFilePath}} && \
echo '<tls-auth>' >> {{ .OVPNFilePath}} && \
cat {{ .KeysDir}}/ta.key >> {{ .OVPNFilePath}} && \
echo '</tls-auth>' >> {{ .OVPNFilePath}}
`

var revokeClientCertCommandTemplate = `
cd {{ .EasyRSADir}} && \
echo yes | ./easyrsa revoke {{ .CName}} && \
./easyrsa gen-crl && \
chmod 755 pki/crl.pem && \
cp pki/crl.pem {{ .KeysDir}}
`

var serverConfigTemplate = `
port {{ .Port}}
proto {{ .Protocol}}
dev tun
ca {{ .KeysDir}}/ca.crt
cert {{ .KeysDir}}/server.crt
key {{ .KeysDir}}/server.key
dh {{ .KeysDir}}/dh.pem
server 10.8.0.0 255.255.255.0
push "redirect-gateway def1 bypass-dhcp"
push "dhcp-option DNS 208.67.222.222"
push "dhcp-option DNS 208.67.220.220"
keepalive 10 120
tls-auth {{ .KeysDir}}/ta.key 0
cipher {{ .Encryption}}
comp-lzo
user nobody
group nogroup
persist-key
persist-tun
status {{ .StatusLogFilePath}} 2
verb 3

auth SHA256
crl-verify {{ .KeysDir}}/crl.pem
key-direction 0
management 127.0.0.1 {{ .ManagementPort}}
ncp-disable
tls-cipher TLS-ECDHE-RSA-WITH-AES-256-GCM-SHA384:TLS-ECDHE-RSA-WITH-AES-256-GCM-SHA384:TLS-DHE-RSA-WITH-AES-256-GCM-SHA384
tls-version-min 1.2
`

var clientConfigTemplate = `
client
dev tun
proto {{ .Protocol}}
remote {{ .IPAddress}} {{ .Port}}
resolv-retry infinite
nobind
user nobody
group nogroup
persist-key
persist-tun
remote-cert-tls server
cipher {{ .Encryption}}
comp-lzo
verb 3
auth SHA256
key-direction 1
script-security 2
`

var firewallRulesTemplate = `
echo 'net.ipv4.ip_forward = 1' >> {{ .SysctlPath}} && \
sysctl -p {{ .SysctlPath}} && \

sed -i '1s@^@\n@g' {{ .UFWRulesPath}} && \
sed -i '1s@^@COMMIT\n@g' {{ .UFWRulesPath}} && \
sed -i '1s@^@-A POSTROUTING -s 10.8.0.0/8 -o eth0 -j MASQUERADE\n@g' {{ .UFWRulesPath}} && \
sed -i '1s@^@:POSTROUTING ACCEPT [0:0]\n@g' {{ .UFWRulesPath}} && \
sed -i '1s@^@*nat\n@g' {{ .UFWRulesPath}} && \
sed -i 's@DEFAULT_FORWARD_POLICY@DEFAULT_FORWARD_POLICY="ACCEPT"\n# DEFAULT_FORWARD_POLICY@g' {{ .UFWDefaultPath}} && \
sed -i 's@DEFAULT_INPUT_POLICY@DEFAULT_INPUT_POLICY="ACCEPT"\n# DEFAULT_INPUT_POLICY@g' {{ .UFWDefaultPath}};
`