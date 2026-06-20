package config

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/sentinel-official/sentinel-go-sdk/libs/netip"
	"github.com/sentinel-official/sentinel-go-sdk/types"
	"github.com/sentinel-official/sentinel-go-sdk/utils"
	"github.com/sentinel-official/sentinelhub/v12/types/v1"
	"github.com/spf13/pflag"
)

const MaxRemoteAddrLen = (1 << 6) - 1 // Maximum allowable length for a remote address.

// allowedServiceTypeSet returns the set of service type strings accepted in service_types.
func allowedServiceTypeSet() map[string]bool {
	return map[string]bool{
		types.ServiceTypeV2Ray.String():     true,
		types.ServiceTypeWireGuard.String(): true,
		types.ServiceTypeOpenVPN.String():   true,
		types.ServiceTypeAmneziaWG.String(): true,
		types.ServiceTypeHysteria2.String(): true,
		types.ServiceTypeXray.String():      true,
	}
}

type NodeConfig struct {
	APIPort                                string   `mapstructure:"api_port"`                                    // APIPort is the port for API access.
	GigabytePrices                         string   `mapstructure:"gigabyte_prices"`                             // GigabytePrices is the pricing information for gigabytes.
	HourlyPrices                           string   `mapstructure:"hourly_prices"`                               // HourlyPrices is the pricing information for hourly usage.
	IntervalBestRPCAddr                    string   `mapstructure:"interval_best_rpc_addr"`                      // IntervalBestRPCAddr is the duration between checking the best RPC address.
	IntervalGeoIPLocation                  string   `mapstructure:"interval_geoip_location"`                     // IntervalGeoIPLocation is the duration between checking the GeoIP location.
	IntervalPricesUpdate                   string   `mapstructure:"interval_prices_update"`                      // IntervalPricesUpdate is the duration between updating the prices of the node.
	IntervalSessionUsageSyncWithBlockchain string   `mapstructure:"interval_session_usage_sync_with_blockchain"` // IntervalSessionUsageSyncWithBlockchain is the duration between syncing session usage with the blockchain.
	IntervalSessionUsageSyncWithDatabase   string   `mapstructure:"interval_session_usage_sync_with_database"`   // IntervalSessionUsageSyncWithDatabase is the duration between syncing session usage with the database.
	IntervalSessionUsageValidate           string   `mapstructure:"interval_session_usage_validate"`             // IntervalSessionUsageValidate is the duration between validating session usage.
	IntervalSessionValidate                string   `mapstructure:"interval_session_validate"`                   // IntervalSessionValidate is the duration between validating sessions.
	IntervalSpeedtest                      string   `mapstructure:"interval_speedtest"`                          // IntervalSpeedtest is the duration between performing speed tests.
	IntervalStatusUpdate                   string   `mapstructure:"interval_status_update"`                      // IntervalStatusUpdate is the duration between updating the status of the node.
	Moniker                                string   `mapstructure:"moniker"`                                     // Moniker is the name or identifier for the node.
	RemoteAddrs                            []string `mapstructure:"remote_addrs"`                                // RemoteAddrs is a list of remote addresses for operations.
	ServiceTypes                           []string `mapstructure:"service_types"`                               // ServiceTypes is the list of enabled service protocols.
	SkipFailedServices                     bool     `mapstructure:"skip_failed_services"`                        // SkipFailedServices skips services that fail to start instead of stopping the node.
}

// APIAddrs generates the API addresses for the node.
func (c *NodeConfig) APIAddrs() []string {
	addrs := make([]string, len(c.RemoteAddrs))
	port := strconv.FormatUint(uint64(c.GetAPIPort().OutFrom), 10)

	for i, addr := range c.RemoteAddrs {
		addrs[i] = net.JoinHostPort(addr, port)
	}

	return addrs
}

// APIListenAddr returns the API listen address.
func (c *NodeConfig) APIListenAddr() string {
	return fmt.Sprintf(":%d", c.APIListenPort())
}

// APIListenPort returns the API listen port.
func (c *NodeConfig) APIListenPort() uint16 {
	return c.GetAPIPort().InFrom
}

// GetAPIPort returns the APIPort field.
func (c *NodeConfig) GetAPIPort() *netip.Port {
	v, err := netip.NewPortFromString(c.APIPort)
	if err != nil {
		panic(err)
	}

	return v
}

// GetGigabytePrices returns the GigabytePrices field.
func (c *NodeConfig) GetGigabytePrices() v1.Prices {
	v, err := v1.NewPricesFromString(c.GigabytePrices)
	if err != nil {
		panic(err)
	}

	return v
}

// GetHourlyPrices returns the HourlyPrices field.
func (c *NodeConfig) GetHourlyPrices() v1.Prices {
	v, err := v1.NewPricesFromString(c.HourlyPrices)
	if err != nil {
		panic(err)
	}

	return v
}

// GetIntervalBestRPCAddr returns the IntervalBestRPCAddr field.
func (c *NodeConfig) GetIntervalBestRPCAddr() time.Duration {
	v, err := time.ParseDuration(c.IntervalBestRPCAddr)
	if err != nil {
		panic(err)
	}

	return v
}

// GetIntervalGeoIPLocation returns the IntervalGeoIPLocation field.
func (c *NodeConfig) GetIntervalGeoIPLocation() time.Duration {
	v, err := time.ParseDuration(c.IntervalGeoIPLocation)
	if err != nil {
		panic(err)
	}

	return v
}

// GetIntervalPricesUpdate returns the IntervalPricesUpdate field.
func (c *NodeConfig) GetIntervalPricesUpdate() time.Duration {
	v, err := time.ParseDuration(c.IntervalPricesUpdate)
	if err != nil {
		panic(err)
	}

	return v
}

// GetIntervalSessionUsageSyncWithBlockchain returns the IntervalSessionUsageSyncWithBlockchain field.
func (c *NodeConfig) GetIntervalSessionUsageSyncWithBlockchain() time.Duration {
	v, err := time.ParseDuration(c.IntervalSessionUsageSyncWithBlockchain)
	if err != nil {
		panic(err)
	}

	return v
}

// GetIntervalSessionUsageSyncWithDatabase returns the IntervalSessionUsageSyncWithDatabase field.
func (c *NodeConfig) GetIntervalSessionUsageSyncWithDatabase() time.Duration {
	v, err := time.ParseDuration(c.IntervalSessionUsageSyncWithDatabase)
	if err != nil {
		panic(err)
	}

	return v
}

// GetIntervalSessionUsageValidate returns the IntervalSessionUsageValidate field.
func (c *NodeConfig) GetIntervalSessionUsageValidate() time.Duration {
	v, err := time.ParseDuration(c.IntervalSessionUsageValidate)
	if err != nil {
		panic(err)
	}

	return v
}

// GetIntervalSessionValidate returns the IntervalSessionValidate field.
func (c *NodeConfig) GetIntervalSessionValidate() time.Duration {
	v, err := time.ParseDuration(c.IntervalSessionValidate)
	if err != nil {
		panic(err)
	}

	return v
}

// GetIntervalSpeedtest returns the IntervalSpeedtest field.
func (c *NodeConfig) GetIntervalSpeedtest() time.Duration {
	v, err := time.ParseDuration(c.IntervalSpeedtest)
	if err != nil {
		panic(err)
	}

	return v
}

// GetIntervalStatusUpdate returns the IntervalStatusUpdate field.
func (c *NodeConfig) GetIntervalStatusUpdate() time.Duration {
	v, err := time.ParseDuration(c.IntervalStatusUpdate)
	if err != nil {
		panic(err)
	}

	return v
}

// GetMoniker returns the Moniker field.
func (c *NodeConfig) GetMoniker() string {
	return c.Moniker
}

// GetRemoteAddrs returns the RemoteAddrs field.
func (c *NodeConfig) GetRemoteAddrs() []string {
	return c.RemoteAddrs
}

// GetServiceTypes returns the enabled service types.
func (c *NodeConfig) GetServiceTypes() []types.ServiceType {
	out := make([]types.ServiceType, len(c.ServiceTypes))
	for i, s := range c.ServiceTypes {
		out[i] = types.ServiceTypeFromString(s)
	}

	return out
}

// GetSkipFailedServices reports whether failed services are skipped at startup.
func (c *NodeConfig) GetSkipFailedServices() bool {
	return c.SkipFailedServices
}

// Validate validates the node configuration.
func (c *NodeConfig) Validate() error {
	// Ensure the API port is not empty and validate it.
	if c.APIPort == "" {
		return errors.New("api_port cannot be empty")
	}

	if _, err := netip.NewPortFromString(c.APIPort); err != nil {
		return fmt.Errorf("parsing api_port %q: %w", c.APIPort, err)
	}

	// Validate the GigabytePrices field.
	if _, err := v1.NewPricesFromString(c.GigabytePrices); err != nil {
		return fmt.Errorf("parsing gigabyte_prices %q: %w", c.GigabytePrices, err)
	}

	// Validate the HourlyPrices field.
	if _, err := v1.NewPricesFromString(c.HourlyPrices); err != nil {
		return fmt.Errorf("parsing hourly_prices %q: %w", c.HourlyPrices, err)
	}

	// Validate interval fields.
	if _, err := time.ParseDuration(c.IntervalBestRPCAddr); err != nil {
		return fmt.Errorf("parsing interval_best_rpc_addr %q: %w", c.IntervalBestRPCAddr, err)
	}

	if _, err := time.ParseDuration(c.IntervalGeoIPLocation); err != nil {
		return fmt.Errorf("parsing interval_geoip_location %q: %w", c.IntervalGeoIPLocation, err)
	}

	if _, err := time.ParseDuration(c.IntervalPricesUpdate); err != nil {
		return fmt.Errorf("parsing interval_prices_update %q: %w", c.IntervalPricesUpdate, err)
	}

	if _, err := time.ParseDuration(c.IntervalSessionUsageSyncWithBlockchain); err != nil {
		return fmt.Errorf("parsing interval_session_usage_sync_with_blockchain %q: %w",
			c.IntervalSessionUsageSyncWithBlockchain, err)
	}

	if _, err := time.ParseDuration(c.IntervalSessionUsageSyncWithDatabase); err != nil {
		return fmt.Errorf("parsing interval_session_usage_sync_with_database %q: %w",
			c.IntervalSessionUsageSyncWithDatabase, err)
	}

	if _, err := time.ParseDuration(c.IntervalSessionUsageValidate); err != nil {
		return fmt.Errorf("parsing interval_session_usage_validate %q: %w", c.IntervalSessionUsageValidate, err)
	}

	if _, err := time.ParseDuration(c.IntervalSessionValidate); err != nil {
		return fmt.Errorf("parsing interval_session_validate %q: %w", c.IntervalSessionValidate, err)
	}

	if _, err := time.ParseDuration(c.IntervalSpeedtest); err != nil {
		return fmt.Errorf("parsing interval_speedtest %q: %w", c.IntervalSpeedtest, err)
	}

	if _, err := time.ParseDuration(c.IntervalStatusUpdate); err != nil {
		return fmt.Errorf("parsing interval_status_update %q: %w", c.IntervalStatusUpdate, err)
	}

	// Ensure the Moniker field is not empty.
	if c.Moniker == "" {
		return errors.New("moniker cannot be empty")
	}

	// Ensure the RemoteAddrs field is not empty.
	if len(c.RemoteAddrs) == 0 {
		return errors.New("remote_addrs cannot be empty")
	}

	// Validate each address in the RemoteAddrs field.
	for _, addr := range c.RemoteAddrs {
		if err := validateRemoteAddr(addr); err != nil {
			return fmt.Errorf("parsing remote_addr %q: %w", addr, err)
		}
	}

	// Validate the service_types list.
	if len(c.ServiceTypes) == 0 {
		return errors.New("service_types cannot be empty")
	}

	allowed := allowedServiceTypeSet()
	seen := make(map[string]bool, len(c.ServiceTypes))

	for _, s := range c.ServiceTypes {
		if !allowed[s] {
			return fmt.Errorf("unsupported service_types entry %q (allowed: amneziawg, hysteria2, openvpn, v2ray, wireguard, xray)", s)
		}

		if seen[s] {
			return fmt.Errorf("duplicate entry %q in service_types", s)
		}

		seen[s] = true
	}

	return nil
}

// SetForFlags adds node configuration flags to the specified FlagSet.
func (c *NodeConfig) SetForFlags(f *pflag.FlagSet) {
	f.StringVar(&c.APIPort, "node.api-port", c.APIPort, "port for API access")
	f.StringVar(&c.GigabytePrices, "node.gigabyte-prices", c.GigabytePrices, "pricing information for gigabytes")
	f.StringVar(&c.HourlyPrices, "node.hourly-prices", c.HourlyPrices, "pricing information for hourly usage")
	f.StringVar(&c.IntervalBestRPCAddr, "node.interval-best-rpc-addr", c.IntervalBestRPCAddr, "interval for checking the best RPC address")
	f.StringVar(&c.IntervalGeoIPLocation, "node.interval-geoip-location", c.IntervalGeoIPLocation, "interval for checking GeoIP location")
	f.StringVar(&c.IntervalPricesUpdate, "node.interval-prices-update", c.IntervalPricesUpdate, "interval for updating node prices")
	f.StringVar(&c.IntervalSessionUsageSyncWithBlockchain, "node.interval-session-usage-sync-with-blockchain", c.IntervalSessionUsageSyncWithBlockchain, "interval for syncing session usage with blockchain")
	f.StringVar(&c.IntervalSessionUsageSyncWithDatabase, "node.interval-session-usage-sync-with-database", c.IntervalSessionUsageSyncWithDatabase, "interval for syncing session usage with database")
	f.StringVar(&c.IntervalSessionUsageValidate, "node.interval-session-usage-validate", c.IntervalSessionUsageValidate, "interval for validating session usage")
	f.StringVar(&c.IntervalSessionValidate, "node.interval-session-validate", c.IntervalSessionValidate, "interval for validating sessions")
	f.StringVar(&c.IntervalSpeedtest, "node.interval-speedtest", c.IntervalSpeedtest, "interval for performing speed tests")
	f.StringVar(&c.IntervalStatusUpdate, "node.interval-status-update", c.IntervalStatusUpdate, "interval for updating node status")
	f.StringVar(&c.Moniker, "node.moniker", c.Moniker, "moniker (identifier) for the node")
	f.StringSliceVar(&c.RemoteAddrs, "node.remote-addrs", c.RemoteAddrs, "list of remote addresses for the node")
	f.StringSliceVar(&c.ServiceTypes, "node.service-types", c.ServiceTypes, "enabled service protocols (e.g. wireguard,amneziawg,hysteria2)")
	f.BoolVar(&c.SkipFailedServices, "node.skip-failed-services", c.SkipFailedServices, "skip services that fail to start instead of stopping the node")
}

// DefaultNodeConfig returns a NodeConfig instance with default values.
func DefaultNodeConfig() *NodeConfig {
	return &NodeConfig{
		APIPort:                                strconv.FormatUint(uint64(utils.RandomPort()), 10),
		GigabytePrices:                         "udvpn:0.0025,12_500_000",
		HourlyPrices:                           "udvpn:0.005,25_000_000",
		IntervalBestRPCAddr:                    (5 * time.Minute).String(),
		IntervalGeoIPLocation:                  (6 * time.Hour).String(),
		IntervalPricesUpdate:                   (6 * time.Hour).String(),
		IntervalSessionUsageSyncWithBlockchain: (2*time.Hour - 5*time.Minute).String(),
		IntervalSessionUsageSyncWithDatabase:   (2 * time.Second).String(),
		IntervalSessionUsageValidate:           (5 * time.Second).String(),
		IntervalSessionValidate:                (5 * time.Minute).String(),
		IntervalSpeedtest:                      (7 * 24 * time.Hour).String(),
		IntervalStatusUpdate:                   (1*time.Hour - 5*time.Minute).String(),
		Moniker:                                randMoniker(),
		RemoteAddrs:                            []string{"127.0.0.1"},
		ServiceTypes:                           randServiceTypes(),
		SkipFailedServices:                     false,
	}
}

// randServiceTypes returns one WireGuard-family service, Hysteria2, and one of
// Xray/V2Ray for a collision-free default set.
func randServiceTypes() []string {
	wireGuardFamily := []string{
		types.ServiceTypeWireGuard.String(),
		types.ServiceTypeAmneziaWG.String(),
	}
	proxyFamily := []string{
		types.ServiceTypeXray.String(),
		types.ServiceTypeV2Ray.String(),
	}

	return []string{
		wireGuardFamily[rand.IntN(len(wireGuardFamily))],
		types.ServiceTypeHysteria2.String(),
		proxyFamily[rand.IntN(len(proxyFamily))],
	}
}

func randMoniker() string {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"

	var result strings.Builder
	result.Grow(8)

	for i := range 8 {
		if i%2 == 0 {
			result.WriteByte(letters[rand.IntN(len(letters))])
		} else {
			result.WriteByte(numbers[rand.IntN(len(numbers))])
		}
	}

	return result.String()
}

func validateRemoteAddr(addr string) error {
	// Ensure the address is not empty or too long.
	if len(addr) == 0 {
		return errors.New("addr cannot be empty")
	}

	if len(addr) > MaxRemoteAddrLen {
		return fmt.Errorf("addr length cannot be greater than %d", MaxRemoteAddrLen)
	}

	// Validate the IP address format.
	if ip := net.ParseIP(addr); ip != nil {
		if ipv4 := ip.To4(); ipv4 != nil {
			return nil
		}

		if ipv6 := ip.To16(); ipv6 != nil {
			return nil
		}

		return errors.New("addr is neither IPv4 nor IPv6")
	}

	// Validate the DNS name format.
	if govalidator.IsDNSName(addr) {
		return nil
	}

	return fmt.Errorf("unsupported addr %q", addr)
}
