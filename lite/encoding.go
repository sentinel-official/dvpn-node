package lite

import (
	sdkstd "github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting"
	hubparams "github.com/sentinel-official/hub/params"
	"github.com/sentinel-official/hub/x/vpn"
)

func EncodingConfig() hubparams.EncodingConfig {
	var (
		cfg     = hubparams.MakeEncodingConfig()
		modules = module.NewBasicManager(
			auth.AppModuleBasic{},
			authvesting.AppModuleBasic{},
			vpn.AppModuleBasic{},
		)
	)

	sdkstd.RegisterLegacyAminoCodec(cfg.Amino)
	sdkstd.RegisterInterfaces(cfg.InterfaceRegistry)
	modules.RegisterLegacyAminoCodec(cfg.Amino)
	modules.RegisterInterfaces(cfg.InterfaceRegistry)

	return cfg
}
