package utils

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	hubtypes "github.com/sentinel-official/hub/types"
	"github.com/showwin/speedtest-go/speedtest"
)

func FindInternetSpeed() (*hubtypes.Bandwidth, error) {
	user, err := speedtest.FetchUserInfo()
	if err != nil {
		return nil, err
	}

	servers, err := speedtest.FetchServers(user)
	if err != nil {
		return nil, err
	}

	var (
		upload   = sdk.ZeroDec()
		download = sdk.ZeroDec()
	)

	for _, s := range servers {
		if s.PingTest() == nil && s.DownloadTest(false) == nil && s.UploadTest(false) == nil {
			upload = sdk.MustNewDecFromStr(fmt.Sprintf("%f", s.ULSpeed))
			download = sdk.MustNewDecFromStr(fmt.Sprintf("%f", s.DLSpeed))
			break
		}
	}

	return &hubtypes.Bandwidth{
		Upload:   upload.Mul(sdk.NewDec(1e6)).QuoInt64(8).TruncateInt(),
		Download: download.Mul(sdk.NewDec(1e6)).QuoInt64(8).TruncateInt(),
	}, nil
}
