package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/sentinel-official/dvpn-node/api"
	"github.com/sentinel-official/dvpn-node/context"
	"github.com/sentinel-official/dvpn-node/libs/geoip"
	"github.com/sentinel-official/dvpn-node/lite"
	"github.com/sentinel-official/dvpn-node/node"
	"github.com/sentinel-official/dvpn-node/services/v2ray"
	"github.com/sentinel-official/dvpn-node/services/wireguard"
	wgtypes "github.com/sentinel-official/dvpn-node/services/wireguard/types"
	"github.com/sentinel-official/dvpn-node/types"
	"github.com/sentinel-official/dvpn-node/utils"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func runHandshake(peers uint64) error {
	return exec.Command("hnsd",
		strings.Split(fmt.Sprintf("--log-file /dev/null "+
			"--pool-size %d "+
			"--rs-host 0.0.0.0:53", peers), " ")...).Run()
}

func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the VPN node",
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
			config, err := types.ReadInConfig(v)
			if err != nil {
				return err
			}

			skipConfigValidation, err := cmd.Flags().GetBool(flagSkipConfigValidation)
			if err != nil {
				return err
			}

			if !skipConfigValidation {
				log.Info("Validating the configuration", "data", config)
				if err = config.Validate(); err != nil {
					return err
				}
			}

			var service types.Service
			if config.Node.Type == "wireguard" {
				log.Info("Creating the IPv4 pool", "CIDR", types.IPv4CIDR)
				ipv4Pool, err := wgtypes.NewIPv4PoolFromCIDR(types.IPv4CIDR)
				if err != nil {
					return err
				}

				log.Info("Creating the IPv6 pool", "CIDR", types.IPv6CIDR)
				ipv6Pool, err := wgtypes.NewIPv6PoolFromCIDR(types.IPv6CIDR)
				if err != nil {
					return err
				}

				service = wireguard.NewWireGuard(wgtypes.NewIPPool(ipv4Pool, ipv6Pool))
			} else if config.Node.Type == "v2ray" {
				service = v2ray.NewV2Ray()
			}

			var (
				input   = bufio.NewReader(cmd.InOrStdin())
				remotes = strings.Split(config.Chain.RPCAddresses, ",")
			)

			log.Info("Initializing the keyring", "name", types.KeyringName, "backend", config.Keyring.Backend)
			kr, err := keyring.New(types.KeyringName, config.Keyring.Backend, home, input)
			if err != nil {
				return err
			}

			info, err := kr.Key(config.Keyring.From)
			if err != nil {
				return err
			}

			client := lite.NewDefaultClient().
				WithChainID(config.Chain.ID).
				WithFromAddress(info.GetAddress()).
				WithFromName(config.Keyring.From).
				WithGas(config.Chain.Gas).
				WithGasAdjustment(config.Chain.GasAdjustment).
				WithGasPrices(config.Chain.GasPrices).
				WithKeyring(kr).
				WithLogger(log).
				WithQueryTimeout(config.Chain.RPCQueryTimeout).
				WithRemotes(remotes).
				WithSignModeStr("").
				WithSimulateAndExecute(config.Chain.SimulateAndExecute).
				WithTxTimeout(config.Chain.RPCTxTimeout)

			account, err := client.QueryAccount(client.FromAddress())
			if err != nil {
				return err
			}
			if account == nil {
				return fmt.Errorf("account does not exist with address %s", client.FromAddress())
			}

			log.Info("Fetching the GeoIP location info...")
			location, err := geoip.Location()
			if err != nil {
				return err
			}
			log.Info("GeoIP location info", "city", location.City, "country", location.Country)

			log.Info("Performing the internet speed test...")
			bandwidth, err := utils.FindInternetSpeed()
			if err != nil {
				return err
			}
			log.Info("Internet speed test result", "data", bandwidth)

			if config.Handshake.Enable {
				go func() {
					for {
						log.Info("Starting the Handshake process...")
						if err := runHandshake(config.Handshake.Peers); err != nil {
							log.Error("handshake process exited unexpectedly", "error", err)
						}
					}
				}()
			}

			log.Info("Initializing the VPN service", "type", service.Type())
			if err = service.Init(home); err != nil {
				return err
			}

			log.Info("Starting the VPN service", "type", service.Type())
			if err = service.Start(); err != nil {
				return err
			}

			log.Info("Opening the database", "path", databasePath)
			database, err := gorm.Open(
				sqlite.Open(databasePath),
				&gorm.Config{
					Logger:      logger.Discard,
					PrepareStmt: false,
				},
			)
			if err != nil {
				return err
			}

			log.Info("Migrating the database models...")
			if err = database.AutoMigrate(&types.Session{}); err != nil {
				return err
			}

			var (
				ctx            = context.NewContext()
				router         = gin.New()
				corsMiddleware = cors.New(
					cors.Config{
						AllowAllOrigins: true,
						AllowMethods: []string{
							http.MethodGet,
							http.MethodPost,
						},
						AllowHeaders: []string{
							types.ContentType,
						},
					},
				)
			)

			router.Use(corsMiddleware)
			api.RegisterRoutes(ctx, router)

			ctx = ctx.WithBandwidth(bandwidth).
				WithClient(client).
				WithConfig(config).
				WithDatabase(database).
				WithHandler(router).
				WithLocation(location).
				WithLogger(log).
				WithService(service)

			n := node.NewNode(ctx)
			if err = n.Initialize(); err != nil {
				return err
			}

			return n.Start(home)
		},
	}

	cmd.Flags().Bool(flagSkipConfigValidation, false, "skip the validation of configuration")

	return cmd
}
