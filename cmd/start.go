package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	cutils "github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/gorilla/mux"
	"github.com/sentinel-official/hub"
	sent "github.com/sentinel-official/hub/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/sentinel-official/dvpn-node/context"
	"github.com/sentinel-official/dvpn-node/lite"
	"github.com/sentinel-official/dvpn-node/node"
	"github.com/sentinel-official/dvpn-node/rest"
	"github.com/sentinel-official/dvpn-node/services/wireguard"
	wgt "github.com/sentinel-official/dvpn-node/services/wireguard/types"
	"github.com/sentinel-official/dvpn-node/types"
	"github.com/sentinel-official/dvpn-node/utils"
)

func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start VPN node",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := cmd.Flags().GetString(types.FlagHome)
			if err != nil {
				return err
			}

			cfgFilePath := filepath.Join(home, types.ConfigFileName)
			if _, err := os.Stat(cfgFilePath); err != nil {
				return fmt.Errorf("config file does not exist at path '%s'", cfgFilePath)
			}

			ipv4Pool, err := wgt.NewIPv4PoolFromCIDR("10.8.0.2/24")
			if err != nil {
				return err
			}

			ipv6Pool, err := wgt.NewIPv6PoolFromCIDR("fd86:ea04:1115::2/120")
			if err != nil {
				return err
			}

			var (
				cfg     = types.NewConfig()
				cdc     = hub.MakeCodec()
				logger  = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
				service = wireguard.NewWireGuard(wgt.NewIPPool(ipv4Pool, ipv6Pool))
			)

			if err := cfg.LoadFromPath(cfgFilePath); err != nil {
				return err
			}
			if err := cfg.Validate(); err != nil {
				return err
			}

			client, err := lite.NewClientFromConfig(cfg)
			if err != nil {
				return err
			}

			client = client.WithCodec(cdc).
				WithTxEncoder(cutils.GetTxEncoder(cdc))

			account, err := client.QueryAccount(client.FromAddress())
			if err != nil {
				return err
			}
			if account == nil {
				return fmt.Errorf("account does not exist with address '%s'", client.FromAddress())
			}

			logger.Info("Initializing the service", "type", service.Type())
			if err := service.Initialize(home); err != nil {
				return err
			}

			logger.Info("Starting the service", "type", service.Type())
			if err := service.Start(); err != nil {
				return err
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
				WithBandwidth(sent.NewBandwidthFromInt64(upload, download))

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
