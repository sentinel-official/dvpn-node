package cmd

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/sentinel-official/sentinel-go-sdk/amneziawg"
	"github.com/sentinel-official/sentinel-go-sdk/hysteria2"
	"github.com/sentinel-official/sentinel-go-sdk/libs/crypto"
	"github.com/sentinel-official/sentinel-go-sdk/libs/log"
	"github.com/sentinel-official/sentinel-go-sdk/openvpn"
	"github.com/sentinel-official/sentinel-go-sdk/types"
	"github.com/sentinel-official/sentinel-go-sdk/utils"
	"github.com/sentinel-official/sentinel-go-sdk/v2ray"
	"github.com/sentinel-official/sentinel-go-sdk/wireguard"
	"github.com/sentinel-official/sentinel-go-sdk/xray"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sentinel-official/sentinel-dvpnx/config"
)

// decollideMaxRetries is the maximum number of attempts to find a non-colliding port or subnet.
const decollideMaxRetries = 16

// decollideSeeds iterates the enabled service set in ascending ServiceType byte order,
// tracking already-accepted host ports and subnets, and re-rolls any colliding service's
// seed port (and subnet, for WG/AWG/OpenVPN) until the collision is resolved or the
// retry budget is exhausted.
func decollideSeeds(set []types.ServiceType, cfg *config.Config) {
	// Sort the set in ascending ServiceType byte order for deterministic processing.
	sorted := make([]types.ServiceType, len(set))
	copy(sorted, set)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	usedPorts := make(map[uint16]bool)
	usedSubnets := make(map[netip.Prefix]bool)

	for _, t := range sorted {
		sc, ok := cfg.Services[t]
		if !ok {
			continue
		}

		switch t {
		case types.ServiceTypeWireGuard:
			c := sc.(*wireguard.ServerConfig)
			// Re-roll port until distinct.
			for try := 0; try < decollideMaxRetries; try++ {
				if !usedPorts[c.OutPort()] {
					break
				}
				c.Port = strconv.FormatUint(uint64(utils.RandomPort()), 10)
			}
			// Re-roll subnet until distinct.
			for try := 0; try < decollideMaxRetries; try++ {
				p, err := netip.ParsePrefix(c.IPv4Addr)
				if err != nil || !usedSubnets[p.Masked()] {
					break
				}
				c.IPv4Addr = fmt.Sprintf("10.%d.%d.1/24", rand.IntN(256), rand.IntN(256))
			}
			usedPorts[c.OutPort()] = true
			if p, err := netip.ParsePrefix(c.IPv4Addr); err == nil {
				usedSubnets[p.Masked()] = true
			}

		case types.ServiceTypeAmneziaWG:
			c := sc.(*amneziawg.ServerConfig)
			for try := 0; try < decollideMaxRetries; try++ {
				if !usedPorts[c.OutPort()] {
					break
				}
				c.Port = strconv.FormatUint(uint64(utils.RandomPort()), 10)
			}
			for try := 0; try < decollideMaxRetries; try++ {
				p, err := netip.ParsePrefix(c.IPv4Addr)
				if err != nil || !usedSubnets[p.Masked()] {
					break
				}
				c.IPv4Addr = fmt.Sprintf("10.%d.%d.1/24", rand.IntN(256), rand.IntN(256))
			}
			usedPorts[c.OutPort()] = true
			if p, err := netip.ParsePrefix(c.IPv4Addr); err == nil {
				usedSubnets[p.Masked()] = true
			}

		case types.ServiceTypeOpenVPN:
			c := sc.(*openvpn.ServerConfig)
			for try := 0; try < decollideMaxRetries; try++ {
				if !usedPorts[c.OutPort()] {
					break
				}
				c.Port = strconv.FormatUint(uint64(utils.RandomPort()), 10)
			}
			for try := 0; try < decollideMaxRetries; try++ {
				p, err := netip.ParsePrefix(c.IPv4Addr)
				if err != nil || !usedSubnets[p.Masked()] {
					break
				}
				c.IPv4Addr = fmt.Sprintf("10.%d.%d.1/24", rand.IntN(256), rand.IntN(256))
			}
			usedPorts[c.OutPort()] = true
			if p, err := netip.ParsePrefix(c.IPv4Addr); err == nil {
				usedSubnets[p.Masked()] = true
			}

		case types.ServiceTypeHysteria2:
			c := sc.(*hysteria2.ServerConfig)
			// Re-roll Port until not colliding.
			for try := 0; try < decollideMaxRetries; try++ {
				if !usedPorts[c.Port] {
					break
				}
				c.Port = utils.RandomPort()
			}
			// Re-roll AuthPort until not colliding.
			for try := 0; try < decollideMaxRetries; try++ {
				if !usedPorts[c.AuthPort] {
					break
				}
				c.AuthPort = utils.RandomPort()
			}
			// Re-roll StatsPort until not colliding.
			for try := 0; try < decollideMaxRetries; try++ {
				if !usedPorts[c.StatsPort] {
					break
				}
				c.StatsPort = utils.RandomPort()
			}
			usedPorts[c.Port] = true
			usedPorts[c.AuthPort] = true
			usedPorts[c.StatsPort] = true

		case types.ServiceTypeV2Ray:
			c := sc.(*v2ray.ServerConfig)
			for _, inb := range c.Inbounds {
				port := inb.GetPort().OutFrom
				for try := 0; try < decollideMaxRetries; try++ {
					if !usedPorts[port] {
						break
					}
					port = utils.RandomPort()
				}
				inb.Port = strconv.FormatUint(uint64(port), 10)
				usedPorts[port] = true
			}

		case types.ServiceTypeXray:
			c := sc.(*xray.ServerConfig)
			for _, inb := range c.Inbounds {
				port := inb.GetPort().OutFrom
				for try := 0; try < decollideMaxRetries; try++ {
					if !usedPorts[port] {
						break
					}
					port = utils.RandomPort()
				}
				inb.Port = strconv.FormatUint(uint64(port), 10)
				usedPorts[port] = true
			}
		}
	}
}

// buildServerForInit constructs the appropriate ServerService for a given ServiceType
// without calling Setup — Init only writes config files, so Setup is not needed here.
func buildServerForInit(t types.ServiceType, homeDir string, cfg *config.Config) (types.ServerService, error) {
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

// NewInitCmd creates and returns a new Cobra command for initializing the application configuration.
func NewInitCmd(cfg *config.Config) *cobra.Command {
	// Initialize default server configs for all supported services
	cfg.Services = map[types.ServiceType]types.ServiceConfig{
		types.ServiceTypeAmneziaWG: amneziawg.DefaultServerConfig(),
		types.ServiceTypeHysteria2: hysteria2.DefaultServerConfig(),
		types.ServiceTypeOpenVPN:   openvpn.DefaultServerConfig(),
		types.ServiceTypeV2Ray:     v2ray.DefaultServerConfig(),
		types.ServiceTypeWireGuard: wireguard.DefaultServerConfig(),
		types.ServiceTypeXray:      xray.DefaultServerConfig(),
	}

	// Declare variables for CLI flags
	var (
		force       bool
		skipTLS     bool
		skipService bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the application configuration",
		Long: `Creates the application home directory and generates a default config.toml file.
If a configuration file already exists, this command will abort unless the "force" flag
is set to overwrite the existing configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Coerce legacy service_type scalar to service_types list before reading the enabled set.
			cfg.Node.NormalizeServiceTypes()

			// Create the home directory if it doesn't exist
			homeDir := viper.GetString("home")
			if err := os.MkdirAll(homeDir, 0700); err != nil {
				return fmt.Errorf("creating application directory %q: %w", homeDir, err)
			}

			// Construct the full path to the config file
			cfgFile := filepath.Join(homeDir, "config.toml")

			// Check if the config file exists at the specified path
			exists, err := utils.IsFileExists(cfgFile)
			if err != nil {
				return fmt.Errorf("checking if config file %q exists: %w", cfgFile, err)
			}

			// Write default config only if file doesn't exist or force flag is set
			if !exists || force {
				log.Info("Writing app config", "file", cfgFile)

				if err := cfg.WriteAppConfig(cfgFile); err != nil {
					return fmt.Errorf("writing config file %q: %w", cfgFile, err)
				}
			}

			// Generate TLS keys if "skipTLS" is disabled
			if !skipTLS {
				log.Info("Initializing PKI with CA certificate and key", "dir", homeDir)

				pki := crypto.NewPKI(homeDir)
				if err := pki.Init(); err != nil {
					return fmt.Errorf("initializing PKI: %w", err)
				}

				opts := []crypto.CertOption{
					crypto.CertSAN(cfg.Node.GetRemoteAddrs()...),
				}

				log.Info("Issuing certificate and key", "name", "tls")

				if _, _, err := pki.Issue("tls", opts...); err != nil {
					return fmt.Errorf("issuing TLS certificate and key: %w", err)
				}
			}

			// Initialize each enabled service config if "skipService" is disabled
			if !skipService {
				set := cfg.Node.GetServiceTypes()
				decollideSeeds(set, cfg)

				for _, t := range set {
					log.Info("Initializing service", "type", t, "force", force)

					svc, err := buildServerForInit(t, homeDir, cfg)
					if err != nil {
						return fmt.Errorf("building service %q for init: %w", t, err)
					}

					if err := svc.Init(force); err != nil {
						return fmt.Errorf("running service init task for %q: %w", t, err)
					}
				}
			}

			log.Info("Configuration initialized successfully")

			return nil
		},
	}

	// Set CLI flags for application and service configuration
	cfg.SetForFlags(cmd.Flags())
	cfg.Services[types.ServiceTypeAmneziaWG].SetForFlags(cmd.Flags(), "amneziawg")
	cfg.Services[types.ServiceTypeHysteria2].SetForFlags(cmd.Flags(), "hysteria2")
	cfg.Services[types.ServiceTypeOpenVPN].SetForFlags(cmd.Flags(), "openvpn")
	cfg.Services[types.ServiceTypeV2Ray].SetForFlags(cmd.Flags(), "v2ray")
	cfg.Services[types.ServiceTypeWireGuard].SetForFlags(cmd.Flags(), "wireguard")
	cfg.Services[types.ServiceTypeXray].SetForFlags(cmd.Flags(), "xray")

	// Bind command-line flags to local variables
	cmd.Flags().BoolVar(&force, "force", force, "overwrite the existing configuration file if it exists")
	cmd.Flags().BoolVar(&skipTLS, "skip-tls", false, "skip TLS key and certificate generation")
	cmd.Flags().BoolVar(&skipService, "skip-service", false, "skip initialization of the selected service")

	_ = cmd.MarkFlagRequired("node.remote-addrs")

	return cmd
}
