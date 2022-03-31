package utils

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	hubtypes "github.com/sentinel-official/hub/types"
	"github.com/showwin/speedtest-go/speedtest"
)

func Bandwidth() (*hubtypes.Bandwidth, error) {
	user, err := speedtest.FetchUserInfo()
	if err != nil {
		return nil, err
	}

	list, err := speedtest.FetchServers(user)
	if err != nil {
		return nil, err
	}

	targets, err := list.FindServer(nil)
	if err != nil {
		return nil, err
	}

	var (
		upload   = sdk.ZeroDec()
		download = sdk.ZeroDec()
	)

	for _, target := range targets {
		if err := target.PingTest(); err != nil {
			return nil, err
		}
		if err := target.DownloadTest(false); err != nil {
			return nil, err
		}
		if err := target.UploadTest(false); err != nil {
			return nil, err
		}

		upload = upload.Add(sdk.MustNewDecFromStr(fmt.Sprintf("%f", target.ULSpeed)))
		download = download.Add(sdk.MustNewDecFromStr(fmt.Sprintf("%f", target.DLSpeed)))
	}

	return &hubtypes.Bandwidth{
		Upload:   upload.Mul(sdk.NewDec(1e6)).QuoInt64(8).TruncateInt(),
		Download: download.Mul(sdk.NewDec(1e6)).QuoInt64(8).TruncateInt(),
	}, nil
}
