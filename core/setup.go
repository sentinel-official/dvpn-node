package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/sentinel-official/sentinel-go-sdk/amneziawg"
	"github.com/sentinel-official/sentinel-go-sdk/core"
	"github.com/sentinel-official/sentinel-go-sdk/hysteria2"
	"github.com/sentinel-official/sentinel-go-sdk/libs/geoip"
	"github.com/sentinel-official/sentinel-go-sdk/libs/log"
	"github.com/sentinel-official/sentinel-go-sdk/libs/oracle"
	"github.com/sentinel-official/sentinel-go-sdk/openvpn"
	"github.com/sentinel-official/sentinel-go-sdk/types"
	"github.com/sentinel-official/sentinel-go-sdk/v2ray"
	"github.com/sentinel-official/sentinel-go-sdk/wireguard"
	"github.com/sentinel-official/sentinel-go-sdk/xray"

	"github.com/sentinel-official/sentinel-dvpnx/config"
	"github.com/sentinel-official/sentinel-dvpnx/database"
)

// SetupAccAddr retrieves the account address for transactions and assigns it to the context.
func (c *Context) SetupAccAddr(ctx context.Context, cfg *config.Config) error {
	log.Info("Retrieving addr for key", "name", cfg.Tx.GetFromName())

	addr, err := c.Client().KeyAddr(cfg.Tx.GetFromName())
	if err != nil {
		return fmt.Errorf("getting addr for key %q: %w", cfg.Tx.GetFromName(), err)
	}

	log.Info("Querying account information", "addr", addr)

	acc, err := c.Client().Account(ctx, addr)
	if err != nil {
		return fmt.Errorf("querying account %q: %w", addr, err)
	}

	if acc == nil {
		return fmt.Errorf("account %s does not exist", addr)
	}

	// Assign the account address to the context.
	c.WithAccAddr(addr)

	return nil
}

// SetupClient initializes the SDK client with the given configuration and assigns it to the context.
func (c *Context) SetupClient(cfg *config.Config) error {
	log.Info("Initializing blockchain client",
		"keyring.backend", cfg.Keyring.GetBackend(),
		"keyring.name", cfg.Keyring.GetName(),
		"rpc.addr", cfg.RPC.GetAddr(),
		"rpc.chain_id", cfg.RPC.GetChainID(),
		"tx.from_name", cfg.Tx.GetFromName(),
	)

	v, err := core.NewClientFromConfig(cfg.Config)
	if err != nil {
		return fmt.Errorf("creating client from config: %w", err)
	}

	// Seal the client.
	v.Seal()

	// Assign the initialized client to the context.
	c.WithClient(v)

	return nil
}

// SetupDatabase creates and configures the database, then assigns it to the context.
func (c *Context) SetupDatabase(_ *config.Config) error {
	log.Info("Initializing database", "file", c.DatabaseFile())

	db, err := database.NewDefault(c.DatabaseFile())
	if err != nil {
		return fmt.Errorf("initializing database %q: %w", c.DatabaseFile(), err)
	}

	// Assign the database instance to the context.
	c.WithDatabase(db)

	return nil
}

// SetupGeoIPClient initializes the GeoIP client and assigns it to the context.
func (c *Context) SetupGeoIPClient(_ *config.Config) error {
	log.Info("Initializing GeoIP client")

	v := geoip.NewDefaultClient()

	// Assign the GeoIP client to the context.
	c.WithGeoIPClient(v)

	return nil
}

// SetupOracleClient initializes the oracle client and assigns it to the context.
func (c *Context) SetupOracleClient(cfg *config.Config) error {
	var (
		client oracle.Client
		name   = cfg.Oracle.GetName()
	)

	if name == "" {
		return nil
	}

	log.Info("Initializing oracle client", "name", name)

	switch name {
	case "coingecko":
		client = oracle.NewCoinGeckoClient(cfg.Oracle.CoinGecko.GetAPIKey())
	case "osmosis":
		client = oracle.NewOsmosisClient(cfg.Oracle.Osmosis.GetAPIAddr())
	default:
		return fmt.Errorf("unsupported name %q", name)
	}

	// Assign the oracle client to the context.
	c.WithOracleClient(client)

	return nil
}

// buildService constructs the server service for the given type using the
// provided home directory and configuration.
func buildService(t types.ServiceType, homeDir string, cfg *config.Config) (types.ServerService, error) {
	switch t {
	case types.ServiceTypeV2Ray:
		return v2ray.NewServer("v2ray", homeDir, cfg.Services[types.ServiceTypeV2Ray].(*v2ray.ServerConfig)), nil
	case types.ServiceTypeWireGuard:
		return wireguard.NewServer("wireguard", homeDir, cfg.Services[types.ServiceTypeWireGuard].(*wireguard.ServerConfig)), nil
	case types.ServiceTypeOpenVPN:
		return openvpn.NewServer("openvpn", homeDir, cfg.Services[types.ServiceTypeOpenVPN].(*openvpn.ServerConfig)), nil
	case types.ServiceTypeAmneziaWG:
		return amneziawg.NewServer("amneziawg", homeDir, cfg.Services[types.ServiceTypeAmneziaWG].(*amneziawg.ServerConfig)), nil
	case types.ServiceTypeHysteria2:
		return hysteria2.NewServer("hysteria2", homeDir, cfg.Services[types.ServiceTypeHysteria2].(*hysteria2.ServerConfig)), nil
	case types.ServiceTypeXray:
		return xray.NewServer("xray", homeDir, cfg.Services[types.ServiceTypeXray].(*xray.ServerConfig)), nil
	case types.ServiceTypeUnspecified:
		return nil, errors.New("unspecified service type")
	default:
		return nil, fmt.Errorf("unsupported service type %q", t)
	}
}

// assembleServices builds and sets up each service in set best-effort, applies
// collision filtering from the conflicts map, and returns the surviving services.
// Returns an error only if zero services could be started.
func assembleServices(
	ctx context.Context,
	set []types.ServiceType,
	conflicts map[types.ServiceType]error,
	build func(types.ServiceType) (types.ServerService, error),
) (map[types.ServiceType]types.ServerService, error) {
	result := make(map[types.ServiceType]types.ServerService, len(set))

	for _, t := range set {
		svc, err := build(t)
		if err != nil {
			log.Warn("Skipping service: build failed", "type", t, "error", err)
			continue
		}

		ok, err := svc.IsRunning()
		if err != nil {
			log.Warn("Skipping service: status check failed", "type", t, "error", err)
			continue
		}

		if ok {
			log.Warn("Skipping service: already running", "type", t)
			continue
		}

		if err := svc.Setup(ctx); err != nil {
			log.Warn("Skipping service: setup failed", "type", t, "error", err)
			continue
		}

		result[t] = svc
	}

	// Apply collision filtering over the successfully-set-up services.
	for t, collErr := range conflicts {
		if _, present := result[t]; present {
			log.Warn("Skipping service: collision detected", "type", t, "error", collErr)
			delete(result, t)
		}
	}

	if len(result) == 0 {
		return nil, errors.New("no services could be started")
	}

	return result, nil
}

// SetupServices builds and sets up each enabled service from the configuration
// best-effort, detects port/subnet collisions among successfully-set-up services,
// and stores the surviving services in the context.  Returns an error only if
// zero services start successfully.
func (c *Context) SetupServices(ctx context.Context, cfg *config.Config) error {
	set := cfg.Node.GetServiceTypes()

	services, err := assembleServices(ctx, set, nil, func(t types.ServiceType) (types.ServerService, error) {
		return buildService(t, c.HomeDir(), cfg)
	})
	if err != nil {
		return err //nolint:wrapcheck
	}

	// Compute collision map from the post-setup survivors and re-assemble.
	setOK := make([]types.ServiceType, 0, len(services))
	for t := range services {
		setOK = append(setOK, t)
	}

	collisions := conflictingServices(setOK, cfg)
	if len(collisions) > 0 {
		for t, collErr := range collisions {
			log.Warn("Removing service after post-setup collision check", "type", t, "error", collErr)
			delete(services, t)
		}

		if len(services) == 0 {
			return errors.New("no services could be started after collision check")
		}
	}

	c.WithServices(services)

	return nil
}

// Setup initializes all components of the node context.
func (c *Context) Setup(ctx context.Context, cfg *config.Config) error {
	// Assign configuration values to the context.
	c.WithAPIAddrs(cfg.Node.APIAddrs())
	c.WithAPIListenAddr(cfg.Node.APIListenAddr())
	c.WithGigabytePrices(cfg.Node.GetGigabytePrices())
	c.WithHandshakeDNS(cfg.HandshakeDNS.GetEnable())
	c.WithHourlyPrices(cfg.Node.GetHourlyPrices())
	c.WithMaxPeers(cfg.QoS.GetMaxPeers())
	c.WithMoniker(cfg.Node.GetMoniker())
	c.WithRemoteAddrs(cfg.Node.GetRemoteAddrs())
	c.WithRPCAddrs(cfg.RPC.GetAddrs())

	log.Info("Setting up blockchain client")

	if err := c.SetupClient(cfg); err != nil {
		return fmt.Errorf("setting up client: %w", err)
	}

	log.Info("Setting up database")

	if err := c.SetupDatabase(cfg); err != nil {
		return fmt.Errorf("setting up database: %w", err)
	}

	log.Info("Setting up GeoIP client")

	if err := c.SetupGeoIPClient(cfg); err != nil {
		return fmt.Errorf("setting up GeoIP client: %w", err)
	}

	log.Info("Setting up oracle client")

	if err := c.SetupOracleClient(cfg); err != nil {
		return fmt.Errorf("setting up oracle client: %w", err)
	}

	log.Info("Setting up services")

	if err := c.SetupServices(ctx, cfg); err != nil {
		return fmt.Errorf("setting up services: %w", err)
	}

	log.Info("Setting up account addr")

	if err := c.SetupAccAddr(ctx, cfg); err != nil {
		return fmt.Errorf("setting up account addr: %w", err)
	}

	return nil
}
