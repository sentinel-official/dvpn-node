// +build linux
// +build !openwrt

package types

import (
	randutils "github.com/sentinel-official/dvpn-node/utils/rand"
)

func (c *Config) WithDefaultValues() *Config {
	key, err := NewPrivateKey()
	if err != nil {
		panic(err)
	}

	c.IFace = "wg0"
	c.IFaceWAN = "eth0"
	c.IPv4CIDR = IPv4CIDR
	c.IPv6CIDR = IPv6CIDR
	c.ListenPort = randutils.RandomPort()
	c.PrivateKey = key.String()

	return c
}
