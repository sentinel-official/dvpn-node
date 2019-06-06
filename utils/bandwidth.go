package utils

import (
	sdk "github.com/ironman0x7b2/sentinel-sdk/types"
)

func CalculateInternetSpeed() (sdk.Bandwidth, error) {
	netSpeed := sdk.NewBandwidthFromInt64(1000000, 1000000)
	return netSpeed, nil
}
