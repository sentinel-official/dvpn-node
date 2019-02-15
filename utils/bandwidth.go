package utils

import (
	"github.com/ironman0x7b2/sentinel-sdk/types"
)

func CalculateInternetSpeed() (types.Bandwidth, error) {
	netSpeed := types.NewBandwidthFromInt64(1000000, 1000000)
	return netSpeed, nil
}
