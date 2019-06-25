package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/sentinel-official/dvpn-node/types"
)

func publicKey(pk interface{}) interface{} {
	switch k := pk.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForPrivateKey(pk interface{}) *pem.Block {
	switch pk := pk.(type) {
	case *rsa.PrivateKey:
		b := x509.MarshalPKCS1PrivateKey(pk)

		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: b}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(pk)
		if err != nil {
			panic(err)
		}

		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}

// nolint:gocyclo
func NewTLSKey(ip net.IP) error {
	pk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	number, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	now := time.Now()
	certificate := &x509.Certificate{
		SerialNumber: number,
		Subject: pkix.Name{
			Organization: []string{"Sentinel VPN Node"},
		},
		NotBefore:             now,
		NotAfter:              now.Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{ip},
	}

	cert, err := x509.CreateCertificate(rand.Reader, certificate, certificate, publicKey(pk), pk)
	if err != nil {
		return err
	}

	file, err := os.Create(types.DefaultTLSCertFilePath)
	if err != nil {
		return err
	}
	if err = pem.Encode(file, &pem.Block{Type: "CERTIFICATE", Bytes: cert}); err != nil { // nolint:gocritic
		return err
	}
	if err = file.Close(); err != nil { // nolint:gocritic
		return err
	}

	file, err = os.OpenFile(types.DefaultTLSKeyFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	if err = pem.Encode(file, pemBlockForPrivateKey(pk)); err != nil { // nolint:gocritic
		return err
	}
	if err = file.Close(); err != nil { // nolint:gocritic
		return err
	}

	return nil
}
