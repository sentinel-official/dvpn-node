package core

import (
	"context"
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

// SetupServices builds and sets up each enabled service. A service that fails
// is fatal unless skip-failed-services is set, in which case it is logged and skipped.
func (c *Context) SetupServices(ctx context.Context, cfg *config.Config) error {
	set := cfg.Node.GetServiceTypes()
	services := make(map[types.ServiceType]types.ServerService, len(set))

	builders := map[types.ServiceType]func() types.ServerService{
		types.ServiceTypeV2Ray: func() types.ServerService {
			return v2ray.NewServer("v2ray", c.HomeDir(), cfg.Services[types.ServiceTypeV2Ray].(*v2ray.ServerConfig))
		},
		types.ServiceTypeWireGuard: func() types.ServerService {
			return wireguard.NewServer("wireguard", c.HomeDir(), cfg.Services[types.ServiceTypeWireGuard].(*wireguard.ServerConfig))
		},
		types.ServiceTypeOpenVPN: func() types.ServerService {
			return openvpn.NewServer("openvpn", c.HomeDir(), cfg.Services[types.ServiceTypeOpenVPN].(*openvpn.ServerConfig))
		},
		types.ServiceTypeAmneziaWG: func() types.ServerService {
			return amneziawg.NewServer("amneziawg", c.HomeDir(), cfg.Services[types.ServiceTypeAmneziaWG].(*amneziawg.ServerConfig))
		},
		types.ServiceTypeHysteria2: func() types.ServerService {
			return hysteria2.NewServer("hysteria2", c.HomeDir(), cfg.Services[types.ServiceTypeHysteria2].(*hysteria2.ServerConfig))
		},
		types.ServiceTypeXray: func() types.ServerService {
			return xray.NewServer("xray", c.HomeDir(), cfg.Services[types.ServiceTypeXray].(*xray.ServerConfig))
		},
	}

	for _, t := range set {
		build, ok := builders[t]
		if !ok {
			return fmt.Errorf("unsupported service type %q", t)
		}

		service := build()
		if err := service.Setup(ctx); err != nil {
			if c.SkipFailedServices() {
				log.Warn("service did not start", "type", t, "error", err)

				continue
			}

			return fmt.Errorf("setting up service %q: %w", t, err)
		}

		services[t] = service
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
	c.WithSkipFailedServices(cfg.Node.GetSkipFailedServices())

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
