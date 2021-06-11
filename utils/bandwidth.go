package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	hubtypes "github.com/sentinel-official/hub/types"
	"github.com/showwin/speedtest-go/speedtest"
)

func Bandwidth() (*hubtypes.Bandwidth, error) {
	user, err := speedtest.FetchUserInfo()
	if err != nil {
		return nil, err
	}

	list, err := speedtest.FetchServerList(user)
	if err != nil {
		return nil, err
	}

	targets, err := list.FindServer(nil)
	if err != nil {
		return nil, err
	}

	var (
		upload   int64
		download int64
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

		upload += int64((target.ULSpeed * 1e6) / 8)
		download += int64((target.DLSpeed * 1e6) / 8)
	}

	return &hubtypes.Bandwidth{
		Upload:   sdk.NewInt(upload),
		Download: sdk.NewInt(download),
	}, nil
}
