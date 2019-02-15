package open_vpn

import (
	"bytes"
	"path/filepath"
	"text/template"
)

type generateClientKeyData struct {
	EasyRSADir           string
	KeysDir              string
	ClientConfigFilePath string
	OVPNFilePath         string
	CName                string
}

func newGenerateClientKeyData(cname string) generateClientKeyData {
	return generateClientKeyData{
		EasyRSADir:           defaultEasyRSADir,
		KeysDir:              defaultKeysDir,
		ClientConfigFilePath: defaultClientConfigFilePath,
		OVPNFilePath:         filepath.Join(defaultKeysDir, cname+ovpnFileExtension),
		CName:                cname,
	}
}

func cmdGenerateClientKey(cname string) (string, error) {
	t, err := template.New("cmd_generate_client_key").Parse(generateClientKeyCommandTemplate)
	if err != nil {
		return "", err
	}

	var stdout bytes.Buffer
	if err := t.Execute(&stdout, newGenerateClientKeyData(cname)); err != nil {
		return "", err
	}

	return stdout.String(), nil
}

type revokeClientCertData struct {
	EasyRSADir string
	KeysDir    string
	CName      string
}

func newRevokeClientCertData(cname string) revokeClientCertData {
	return revokeClientCertData{
		EasyRSADir: defaultEasyRSADir,
		KeysDir:    defaultKeysDir,
		CName:      cname,
	}
}

func cmdRevokeClientCert(cname string) (string, error) {
	t, err := template.New("cmd_revoke_client_cert").Parse(revokeClientCertCommandTemplate)
	if err != nil {
		return "", err
	}

	var stdout bytes.Buffer
	if err := t.Execute(&stdout, newRevokeClientCertData(cname)); err != nil {
		return "", err
	}

	return stdout.String(), nil
}
