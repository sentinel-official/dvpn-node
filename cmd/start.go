package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/gorilla/mux"
	"github.com/sentinel-official/hub"
	"github.com/sentinel-official/hub/params"
	hubtypes "github.com/sentinel-official/hub/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/sentinel-official/dvpn-node/context"
	"github.com/sentinel-official/dvpn-node/lite"
	"github.com/sentinel-official/dvpn-node/node"
	"github.com/sentinel-official/dvpn-node/rest"
	"github.com/sentinel-official/dvpn-node/services/wireguard"
	wgtypes "github.com/sentinel-official/dvpn-node/services/wireguard/types"
	"github.com/sentinel-official/dvpn-node/types"
	"github.com/sentinel-official/dvpn-node/utils"
)

func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start VPN node",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				home = viper.GetString(flags.FlagHome)
				path = filepath.Join(home, types.ConfigFileName)
			)

			cfg := types.NewConfig().WithDefaultValues()
			if err := cfg.LoadFromPath(path); err != nil {
				return err
			}
			if err := cfg.Validate(); err != nil {
				return err
			}

			ipv4Pool, err := wgtypes.NewIPv4PoolFromCIDR("10.8.0.2/24")
			if err != nil {
				return err
			}

			ipv6Pool, err := wgtypes.NewIPv6PoolFromCIDR("fd86:ea04:1115::2/120")
			if err != nil {
				return err
			}

			var (
				encoding = params.MakeEncodingConfig()
				logger   = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
				service  = wireguard.NewWireGuard(wgtypes.NewIPPool(ipv4Pool, ipv6Pool))
				reader   = bufio.NewReader(cmd.InOrStdin())
			)

			std.RegisterInterfaces(encoding.InterfaceRegistry)
			hub.ModuleBasics.RegisterInterfaces(encoding.InterfaceRegistry)

			rpcclient, err := rpchttp.New(cfg.Chain.RPCAddress, "/websocket")
			if err != nil {
				return err
			}

			kr, err := keyring.New(sdk.KeyringServiceName(), cfg.Keyring.Backend, home, reader)
			if err != nil {
				return err
			}

			info, err := kr.Key(cfg.Keyring.From)
			if err != nil {
				return err
			}

			client := lite.NewDefaultClient().
				WithAccountRetriever(authtypes.AccountRetriever{}).
				WithChainID(cfg.Chain.ID).
				WithClient(rpcclient).
				WithFrom(cfg.Keyring.From).
				WithFromAddress(info.GetAddress()).
				WithFromName(cfg.Keyring.From).
				WithGas(cfg.Chain.Gas).
				WithGasAdjustment(cfg.Chain.GasAdjustment).
				WithGasPrices(cfg.Chain.GasPrices).
				WithInterfaceRegistry(encoding.InterfaceRegistry).
				WithKeyring(kr).
				WithLegacyAmino(encoding.Amino).
				WithNodeURI(cfg.Chain.RPCAddress).
				WithSimulateAndExecute(cfg.Chain.SimulateAndExecute).
				WithTxConfig(encoding.TxConfig)

			account, err := client.QueryAccount(info.GetAddress())
			if err != nil {
				return err
			}
			if account == nil {
				return fmt.Errorf("account does not exist with address %s", client.FromAddress())
			}

			logger.Info("Fetching the GeoIP location")
			location, err := utils.FetchGeoIPLocation()
			if err != nil {
				return err
			}

			logger.Info("Calculating the bandwidth")
			upload, download, err := utils.Bandwidth()
			if err != nil {
				return err
			}

			if cfg.Handshake.Enable {
				if err := runHandshakeDaemon(cfg.Handshake.Peers); err != nil {
					return err
				}
			}

			logger.Info("Initializing the service", "type", service.Type())
			if err := service.Init(home); err != nil {
				return err
			}

			logger.Info("Starting the service", "type", service.Type())
			if err := service.Start(); err != nil {
				return err
			}

			var (
				ctx    = context.NewContext()
				router = mux.NewRouter()
			)

			rest.RegisterRoutes(ctx, router)

			ctx = ctx.
				WithLogger(logger).
				WithService(service).
				WithRouter(router).
				WithConfig(cfg).
				WithClient(client).
				WithHome(home).
				WithLocation(location).
				WithSessions(types.NewSessions()).
				WithBandwidth(hubtypes.NewBandwidthFromInt64(upload, download))

			n := node.NewNode(ctx)
			if err := n.Initialize(); err != nil {
				return err
			}

			return n.Start()
		},
	}

	return cmd
}

func runHandshakeDaemon(peers uint64) error {
	return exec.Command("hnsd",
		strings.Split(fmt.Sprintf("--daemon "+
			"--log-file /dev/null "+
			"--pool-size %d "+
			"--rs-host 0.0.0.0:53", peers), " ")...).Start()
}
