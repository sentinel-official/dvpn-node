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

// SetupService determines the service type and configures it accordingly.
func (c *Context) SetupService(ctx context.Context, cfg *config.Config) error {
	var (
		service     types.ServerService         // Interface for the node service
		serviceType = cfg.Node.GetServiceType() // Get the service type from config
	)

	log.Info("Initializing service", "type", serviceType)

	// Initialize the appropriate server service based on the configured type
	switch serviceType {
	case types.ServiceTypeV2Ray:
		service = v2ray.NewServer("v2ray", c.HomeDir(), cfg.Services[types.ServiceTypeV2Ray].(*v2ray.ServerConfig))
	case types.ServiceTypeWireGuard:
		service = wireguard.NewServer("wireguard", c.HomeDir(), cfg.Services[types.ServiceTypeWireGuard].(*wireguard.ServerConfig))
	case types.ServiceTypeOpenVPN:
		service = openvpn.NewServer("openvpn", c.HomeDir(), cfg.Services[types.ServiceTypeOpenVPN].(*openvpn.ServerConfig))
	case types.ServiceTypeAmneziaWG:
		service = amneziawg.NewServer("amneziawg", c.HomeDir(), cfg.Services[types.ServiceTypeAmneziaWG].(*amneziawg.ServerConfig))
	case types.ServiceTypeHysteria2:
		service = hysteria2.NewServer("hysteria2", c.HomeDir(), cfg.Services[types.ServiceTypeHysteria2].(*hysteria2.ServerConfig))
	case types.ServiceTypeXray:
		service = xray.NewServer("xray", c.HomeDir(), cfg.Services[types.ServiceTypeXray].(*xray.ServerConfig))
	case types.ServiceTypeUnspecified:
		return errors.New("unspecified service type")
	default:
		return fmt.Errorf("unsupported service type %q", serviceType)
	}

	log.Info("Checking service status")

	ok, err := service.IsRunning()
	if err != nil {
		return fmt.Errorf("checking service %q status: %w", serviceType, err)
	}

	if ok {
		return fmt.Errorf("service %q is already running", serviceType)
	}

	if err := service.Setup(ctx); err != nil {
		return err //nolint:wrapcheck
	}

	// Assign the service to the context
	c.WithService(service)

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

	log.Info("Setting up service")

	if err := c.SetupService(ctx, cfg); err != nil {
		return fmt.Errorf("setting up service: %w", err)
	}

	log.Info("Setting up account addr")

	if err := c.SetupAccAddr(ctx, cfg); err != nil {
		return fmt.Errorf("setting up account addr: %w", err)
	}

	return nil
}
