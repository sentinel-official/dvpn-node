// this is a total hack.  It is here so that I can work through how to best integrate v2ray. -jacob
// the goal is to start v2ray with either a generic config, or maybe the config from the original docs.
// I guess it is best to have the new node generate a config though, because we don't want to allow aribitrary access.
//
//nolint:unused
package v2ray

import (
	"fmt"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common/cmdarg"
	"github.com/v2fly/v2ray-core/v5/common/errors"
)

var (
	configFiles          cmdarg.Arg
	configDirs           cmdarg.Arg
	configFormat         *string
	configDirRecursively *bool
)

func startV2Ray() (core.Server, error) {
	config, err := core.LoadConfig(*configFormat, configFiles)
	if err != nil {
		if len(configFiles) == 0 {
			err = newError("failed to load config").Base(err)
		} else {
			err = newError(fmt.Sprintf("failed to load config: %s", configFiles)).Base(err)
		}
		return nil, err
	}

	server, err := core.New(config)
	if err != nil {
		return nil, newError("failed to create server").Base(err)
	}

	return server, nil
}

type errPathObjHolder struct{}

func newError(values ...interface{}) *errors.Error {
	return errors.New(values...).WithPathObj(errPathObjHolder{})
}
