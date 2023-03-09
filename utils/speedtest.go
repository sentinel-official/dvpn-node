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

	servers, err = servers.FindServer(nil)
	if err != nil {
		return nil, err
	}

	var (
		upload   = sdk.ZeroDec()
		download = sdk.ZeroDec()
	)

	for _, s := range servers {
		s.Context.Reset()
		if err = s.PingTest(); err != nil {
			continue
		}

		if err = s.DownloadTest(); err != nil {
			continue
		}
		s.Context.Wait()
		if err = s.UploadTest(); err != nil {
			continue
		}

		upload = sdk.MustNewDecFromStr(fmt.Sprintf("%f", s.ULSpeed))
		download = sdk.MustNewDecFromStr(fmt.Sprintf("%f", s.DLSpeed))

		if upload.IsPositive() && download.IsPositive() {
			break
		}
	}

	return &hubtypes.Bandwidth{
		Upload:   upload.Mul(sdk.NewDec(1e6)).QuoInt64(8).TruncateInt(),
		Download: download.Mul(sdk.NewDec(1e6)).QuoInt64(8).TruncateInt(),
	}, nil
}
