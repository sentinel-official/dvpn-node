package types

import (
	"math/big"
	"net"
)

type (
	IPv4 [net.IPv4len]byte
	IPv6 [net.IPv6len]byte
)

func NewIPv4FromIP(ip net.IP) (v4 IPv4) {
	copy(v4[:], ip.To4())
	return v4
}

func (ip IPv4) Bytes() []byte {
	return ip[:]
}

func (ip IPv4) Empty() bool {
	return ip == IPv4{}
}

func (ip IPv4) IP() net.IP {
	return ip.Bytes()
}

func (ip IPv4) Next() IPv4 {
	next := big.NewInt(0)
	next.Add(new(big.Int).SetBytes(ip.Bytes()), big.NewInt(1))
	copy(ip[:], next.Bytes())

	return ip
}

func NewIPv6FromIP(ip net.IP) (v6 IPv6) {
	copy(v6[:], ip.To16())
	return v6
}

func (ip IPv6) Bytes() []byte {
	return ip[:]
}

func (ip IPv6) Empty() bool {
	return ip == IPv6{}
}

func (ip IPv6) IP() net.IP {
	return ip.Bytes()
}

func (ip IPv6) Next() IPv6 {
	next := big.NewInt(0)
	next.Add(new(big.Int).SetBytes(ip.Bytes()), big.NewInt(1))
	copy(ip[:], next.Bytes())

	return ip
}
