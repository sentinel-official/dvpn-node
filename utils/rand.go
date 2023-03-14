package utils

import (
	"crypto/rand"
	"math/big"
)

func RandomPort() uint16 {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<16-1<<10))
	if err != nil {
		panic(err)
	}

	return uint16(n.Int64() + 1<<10)
}
