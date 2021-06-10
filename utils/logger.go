package utils

import (
	"io"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/config"
	tmlog "github.com/tendermint/tendermint/libs/log"
)

func PrepareLogger() (tmlog.Logger, error) {
	var (
		format           = viper.GetString(flags.FlagLogFormat)
		level            = viper.GetString(flags.FlagLogLevel)
		writer io.Writer = os.Stderr
	)

	if format == config.LogFormatPlain {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		}
	}

	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	return &server.ZeroLogWrapper{
		Logger: zerolog.New(writer).
			Level(logLevel).
			With().Timestamp().
			Logger(),
	}, nil
}
