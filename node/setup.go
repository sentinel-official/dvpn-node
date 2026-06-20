package node

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sentinel-official/sentinel-go-sdk/libs/cmux"
	"github.com/sentinel-official/sentinel-go-sdk/libs/cron"
	"github.com/sentinel-official/sentinel-go-sdk/libs/gin/middlewares"
	"github.com/sentinel-official/sentinel-go-sdk/libs/log"

	"github.com/sentinel-official/sentinel-dvpnx/api"
	"github.com/sentinel-official/sentinel-dvpnx/config"
	"github.com/sentinel-official/sentinel-dvpnx/core"
	"github.com/sentinel-official/sentinel-dvpnx/hnsd"
	"github.com/sentinel-official/sentinel-dvpnx/workers"
)

// SetupScheduler sets up the cron scheduler with various workers.
func (n *Node) SetupScheduler(ctx context.Context, cfg *config.Config) error {
	// Define the list of cron workers with their respective handlers and intervals.
	items := []cron.Worker{
		workers.NewBestRPCAddrWorker(n.Context(), cfg.Node.GetIntervalBestRPCAddr()),
		workers.NewGeoIPLocationWorker(n.Context(), cfg.Node.GetIntervalGeoIPLocation()),
		workers.NewNodePricesUpdateWorker(n.Context(), cfg.Node.GetIntervalPricesUpdate()),
		workers.NewNodeStatusUpdateWorker(n.Context(), cfg.Node.GetIntervalStatusUpdate()),
		workers.NewSessionUsageSyncWithBlockchainWorker(n.Context(), cfg.Node.GetIntervalSessionUsageSyncWithBlockchain()),
		workers.NewSessionUsageSyncWithDatabaseWorker(n.Context(), cfg.Node.GetIntervalSessionUsageSyncWithDatabase()),
		workers.NewSessionUsageValidateWorker(n.Context(), cfg.Node.GetIntervalSessionUsageValidate()),
		workers.NewSessionValidateWorker(n.Context(), cfg.Node.GetIntervalSessionValidate()),
		workers.NewSpeedtestWorker(n.Context(), cfg.Node.GetIntervalSpeedtest()),
	}

	log.Info("Initializing scheduler")

	s := cron.NewScheduler("scheduler")
	if err := s.Setup(ctx); err != nil {
		return err //nolint:wrapcheck
	}

	for _, item := range items {
		log.Info("Registering scheduler worker",
			"name", item.Name(), "interval", item.Interval().String(),
		)

		if err := s.Register(item); err != nil {
			return fmt.Errorf("registering scheduler worker %q: %w", item.Name(), err)
		}
	}

	// Attach the configured scheduler to the Node.
	n.WithScheduler(s)

	return nil
}

// SetupServer sets up the API server with necessary middlewares and API routes.
func (n *Node) SetupServer(ctx context.Context, _ *config.Config) error {
	// Sets the Gin mode to ReleaseMode.
	gin.SetMode(gin.ReleaseMode)

	// Define middlewares to be used by the router.
	items := []gin.HandlerFunc{
		cors.New(
			cors.Config{
				AllowAllOrigins: true,
				AllowMethods:    []string{http.MethodGet, http.MethodPost},
			},
		),
		middlewares.RateLimiter(ctx, nil),
	}

	// Create a new Gin router and apply the middlewares.
	router := gin.New()
	router.Use(items...)

	// Register API routes to the router.
	api.RegisterRoutes(n.Context(), router)

	log.Info("Initializing API server")

	s := cmux.NewServer(
		"API-server",
		n.Context().APIListenAddr(),
		n.Context().TLSCertFile(),
		n.Context().TLSKeyFile(),
		router,
	)
	if err := s.Setup(ctx); err != nil {
		return err //nolint:wrapcheck
	}

	// Attach the API server to the Node instance.
	n.WithServer(s)

	return nil
}

// SetupHandshakeDNS constructs and attaches the hnsd daemon when Handshake DNS is enabled.
func (n *Node) SetupHandshakeDNS(ctx context.Context, cfg *config.Config) error {
	if !cfg.HandshakeDNS.GetEnable() {
		return nil
	}

	gateway, err := cfg.ServiceGatewayAddr()
	if err != nil {
		return fmt.Errorf("resolving service gateway addr: %w", err)
	}

	rsHost := net.JoinHostPort(gateway, "53")

	log.Info("Initializing Handshake DNS",
		"rs_host", rsHost,
		"pool_size", cfg.HandshakeDNS.GetPeers(),
		"max_restarts", cfg.HandshakeDNS.GetMaxRestarts(),
	)

	d := hnsd.New("hnsd", rsHost, cfg.HandshakeDNS.GetPeers(), cfg.HandshakeDNS.GetMaxRestarts())
	if err := d.Setup(ctx); err != nil {
		return err //nolint:wrapcheck
	}

	n.WithHandshakeDNS(d)

	return nil
}

// SetupContext sets up the core context.
func (n *Node) SetupContext(ctx context.Context, homeDir string, input io.Reader, cfg *config.Config) error {
	log.Info("Initializing context")

	c := core.NewContext().
		WithHomeDir(homeDir).
		WithInput(input)
	if err := c.Setup(ctx, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	// Seal the context.
	c.Seal()

	// Attach the code context to the Node instance.
	n.WithContext(c)

	return nil
}

// Setup sets up the context, scheduler and API server for the Node.
func (n *Node) Setup(ctx context.Context, homeDir string, input io.Reader, cfg *config.Config) error {
	return n.Manager.Setup(ctx, func() error { //nolint:wrapcheck
		log.Info("Setting up context")

		if err := n.SetupContext(ctx, homeDir, input, cfg); err != nil {
			return fmt.Errorf("setting up context: %w", err)
		}

		log.Info("Setting up scheduler")

		if err := n.SetupScheduler(ctx, cfg); err != nil {
			return fmt.Errorf("setting up scheduler: %w", err)
		}

		log.Info("Setting up API server")

		if err := n.SetupServer(ctx, cfg); err != nil {
			return fmt.Errorf("setting up API server: %w", err)
		}

		log.Info("Setting up Handshake DNS")

		if err := n.SetupHandshakeDNS(ctx, cfg); err != nil {
			return fmt.Errorf("setting up Handshake DNS: %w", err)
		}

		return nil
	})
}
