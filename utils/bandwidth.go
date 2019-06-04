package utils

import (
	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
)

func CalculateInternetSpeed() (sdkTypes.Bandwidth, error) {
	netSpeed := sdkTypes.NewBandwidthFromInt64(1000000, 1000000)
	return netSpeed, nil
}
