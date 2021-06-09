package cmd

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/std"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/gorilla/mux"
	"github.com/sentinel-official/hub"
	"github.com/sentinel-official/hub/params"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			var (
				home = viper.GetString(flags.FlagHome)
				path = filepath.Join(home, types.ConfigFileName)
			)

			logger, err := utils.PrepareLogger()
			if err != nil {
				return err
			}

			v := viper.New()
			v.SetConfigFile(path)

			logger.Info("Reading configuration file", "path", path)
			cfg, err := types.ReadInConfig(v)
			if err != nil {
				return err
			}

			logger.Info("Validating configuration", "data", cfg)
			if err := cfg.Validate(); err != nil {
				return err
			}

			logger.Info("Creating IPv4 pool", "CIDR", types.DefaultIPv4CIDR)
			ipv4Pool, err := wgtypes.NewIPv4PoolFromCIDR(types.DefaultIPv4CIDR)
			if err != nil {
				return err
			}

			logger.Info("Creating IPv6 pool", "CIDR", types.DefaultIPv6CIDR)
			ipv6Pool, err := wgtypes.NewIPv6PoolFromCIDR(types.DefaultIPv6CIDR)
			if err != nil {
				return err
			}

			var (
				encoding = params.MakeEncodingConfig()
				service  = wireguard.NewWireGuard(wgtypes.NewIPPool(ipv4Pool, ipv6Pool))
				reader   = bufio.NewReader(cmd.InOrStdin())
			)

			std.RegisterInterfaces(encoding.InterfaceRegistry)
			hub.ModuleBasics.RegisterInterfaces(encoding.InterfaceRegistry)

			logger.Info("Initializing RPC HTTP client", "address", cfg.Chain.RPCAddress, "endpoint", "/websocket")
			rpcclient, err := rpchttp.New(cfg.Chain.RPCAddress, "/websocket")
			if err != nil {
				return err
			}

			logger.Info("Initializing keyring", "name", types.KeyringName, "backend", cfg.Keyring.Backend)
			kr, err := keyring.New(types.KeyringName, cfg.Keyring.Backend, home, reader)
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
				WithLogger(logger).
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

			logger.Info("Fetching GeoIP location info...")
			location, err := utils.FetchGeoIPLocation()
			if err != nil {
				return err
			}
			logger.Info("GeoIP location info", "city", location.City, "country", location.Country)

			logger.Info("Performing internet speed test...")
			bandwidth, err := utils.Bandwidth()
			if err != nil {
				return err
			}
			logger.Info("Internet speed test result", "data", bandwidth)

			if cfg.Handshake.Enable {
				if err := runHandshakeDaemon(cfg.Handshake.Peers); err != nil {
					return err
				}
			}

			logger.Info("Initializing underlying VPN service", "type", service.Type())
			if err := service.Init(home); err != nil {
				return err
			}

			logger.Info("Starting underlying VPN service", "type", service.Type())
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
				WithLocation(location).
				WithSessions(types.NewSessions()).
				WithBandwidth(bandwidth)

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
