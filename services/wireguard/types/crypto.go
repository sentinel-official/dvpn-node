package types

import (
	crand "crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

const (
	KeyLength = 32
)

type Key [KeyLength]byte

func KeyFromBytes(b []byte) (*Key, error) {
	if len(b) != KeyLength {
		return nil, fmt.Errorf("invalid bytes length")
	}

	var key Key
	copy(key[:], b)

	return &key, nil
}

func KeyFromString(s string) (*Key, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	return KeyFromBytes(bytes)
}

func NewPreSharedKey() (*Key, error) {
	var key Key

	_, err := crand.Read(key[:])
	if err != nil {
		return nil, err
	}

	return &key, nil
}

func NewPrivateKey() (*Key, error) {
	key, err := NewPreSharedKey()
	if err != nil {
		return nil, err
	}

	key[0] &= 248
	key[31] = (key[31] & 127) | 64
	return key, nil
}

func (k *Key) String() string {
	return base64.StdEncoding.EncodeToString(k[:])
}

func (k *Key) IsZero() bool {
	var zeros Key
	return subtle.ConstantTimeCompare(zeros[:], k[:]) == 1
}

func (k *Key) Public() *Key {
	var p [KeyLength]byte
	curve25519.ScalarBaseMult(&p, (*[KeyLength]byte)(k))
	return (*Key)(&p)
}

func (k *Key) Bytes() []byte {
	return k[:]
}
