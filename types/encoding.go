package types

import (
	sdkstd "github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distributionclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	ibcica "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts"
	ibctransfer "github.com/cosmos/ibc-go/v3/modules/apps/transfer"
	ibc "github.com/cosmos/ibc-go/v3/modules/core"
	ibcclientclient "github.com/cosmos/ibc-go/v3/modules/core/02-client/client"
	hubparams "github.com/sentinel-official/hub/params"
	custommint "github.com/sentinel-official/hub/x/mint"
	"github.com/sentinel-official/hub/x/swap"
	"github.com/sentinel-official/hub/x/vpn"
)

var (
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		authvesting.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModule{},
		crisis.AppModuleBasic{},
		distribution.AppModuleBasic{},
		evidence.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		genutil.AppModuleBasic{},
		gov.NewAppModuleBasic(
			distributionclient.ProposalHandler,
			ibcclientclient.UpdateClientProposalHandler,
			ibcclientclient.UpgradeProposalHandler,
			paramsclient.ProposalHandler,
			upgradeclient.ProposalHandler,
			upgradeclient.CancelProposalHandler,
		),
		ibc.AppModuleBasic{},
		ibcica.AppModuleBasic{},
		ibctransfer.AppModuleBasic{},
		mint.AppModuleBasic{},
		params.AppModuleBasic{},
		slashing.AppModuleBasic{},
		staking.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		custommint.AppModule{},
		swap.AppModuleBasic{},
		vpn.AppModuleBasic{},
	)
)

func MakeEncodingConfig() hubparams.EncodingConfig {
	config := hubparams.MakeEncodingConfig()

	sdkstd.RegisterLegacyAminoCodec(config.Amino)
	sdkstd.RegisterInterfaces(config.InterfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(config.Amino)
	ModuleBasics.RegisterInterfaces(config.InterfaceRegistry)

	return config
}
