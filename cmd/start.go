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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

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
				home         = viper.GetString(flags.FlagHome)
				configPath   = filepath.Join(home, types.ConfigFileName)
				databasePath = filepath.Join(home, types.DatabaseFileName)
			)

			log, err := utils.PrepareLogger()
			if err != nil {
				return err
			}

			v := viper.New()
			v.SetConfigFile(configPath)

			log.Info("Reading the configuration file", "path", configPath)
			cfg, err := types.ReadInConfig(v)
			if err != nil {
				return err
			}

			skipConfigValidation, err := cmd.Flags().GetBool(flagSkipConfigValidation)
			if err != nil {
				return err
			}

			if !skipConfigValidation {
				log.Info("Validating the configuration", "data", cfg)
				if err := cfg.Validate(); err != nil {
					return err
				}
			}

			log.Info("Creating IPv4 pool", "CIDR", types.DefaultIPv4CIDR)
			ipv4Pool, err := wgtypes.NewIPv4PoolFromCIDR(types.DefaultIPv4CIDR)
			if err != nil {
				return err
			}

			log.Info("Creating IPv6 pool", "CIDR", types.DefaultIPv6CIDR)
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

			log.Info("Initializing RPC HTTP client", "address", cfg.Chain.RPCAddress, "endpoint", "/websocket")
			rpcclient, err := rpchttp.New(cfg.Chain.RPCAddress, "/websocket")
			if err != nil {
				return err
			}

			log.Info("Initializing keyring", "name", types.KeyringName, "backend", cfg.Keyring.Backend)
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
				WithLogger(log).
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

			log.Info("Fetching GeoIP location info...")
			location, err := utils.FetchGeoIPLocation()
			if err != nil {
				return err
			}
			log.Info("GeoIP location info", "city", location.City, "country", location.Country)

			log.Info("Performing internet speed test...")
			bandwidth, err := utils.Bandwidth()
			if err != nil {
				return err
			}
			log.Info("Internet speed test result", "data", bandwidth)

			if cfg.Handshake.Enable {
				if err := runHandshakeDaemon(cfg.Handshake.Peers); err != nil {
					return err
				}
			}

			log.Info("Initializing VPN service", "type", service.Type())
			if err := service.Init(home); err != nil {
				return err
			}

			log.Info("Starting VPN service", "type", service.Type())
			if err := service.Start(); err != nil {
				return err
			}

			log.Info("Opening the database", "path", databasePath)
			database, err := gorm.Open(sqlite.Open(databasePath))
			if err != nil {
				return err
			}

			log.Info("Migrating database models...")
			if err := database.AutoMigrate(&types.Session{}); err != nil {
				return err
			}

			var (
				ctx    = context.NewContext()
				router = mux.NewRouter()
			)

			rest.RegisterRoutes(ctx, router)

			ctx = ctx.
				WithLogger(log).
				WithService(service).
				WithRouter(router).
				WithConfig(cfg).
				WithClient(client).
				WithLocation(location).
				WithDatabase(database).
				WithBandwidth(bandwidth)

			n := node.NewNode(ctx)
			if err := n.Initialize(); err != nil {
				return err
			}

			return n.Start()
		},
	}

	cmd.Flags().Bool(flagSkipConfigValidation, false, "skip the validation of configuration file")

	return cmd
}

func runHandshakeDaemon(peers uint64) error {
	return exec.Command("hnsd",
		strings.Split(fmt.Sprintf("--daemon "+
			"--log-file /dev/null "+
			"--pool-size %d "+
			"--rs-host 0.0.0.0:53", peers), " ")...).Start()
}
