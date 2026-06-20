package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

// buildServerForInit constructs the appropriate ServerService for a given ServiceType
// without calling Setup — Init only writes config files, so Setup is not needed here.
func buildServerForInit(t types.ServiceType, homeDir string, cfg *config.Config) (types.ServerService, error) {
	switch t {
	case types.ServiceTypeV2Ray:
		sc, ok := cfg.Services[types.ServiceTypeV2Ray].(*v2ray.ServerConfig)
		if !ok {
			return nil, fmt.Errorf("missing or wrong config type for service %q", t)
		}

		return v2ray.NewServer("v2ray", homeDir, sc), nil
	case types.ServiceTypeWireGuard:
		sc, ok := cfg.Services[types.ServiceTypeWireGuard].(*wireguard.ServerConfig)
		if !ok {
			return nil, fmt.Errorf("missing or wrong config type for service %q", t)
		}

		return wireguard.NewServer("wireguard", homeDir, sc), nil
	case types.ServiceTypeOpenVPN:
		sc, ok := cfg.Services[types.ServiceTypeOpenVPN].(*openvpn.ServerConfig)
		if !ok {
			return nil, fmt.Errorf("missing or wrong config type for service %q", t)
		}

		return openvpn.NewServer("openvpn", homeDir, sc), nil
	case types.ServiceTypeAmneziaWG:
		sc, ok := cfg.Services[types.ServiceTypeAmneziaWG].(*amneziawg.ServerConfig)
		if !ok {
			return nil, fmt.Errorf("missing or wrong config type for service %q", t)
		}

		return amneziawg.NewServer("amneziawg", homeDir, sc), nil
	case types.ServiceTypeHysteria2:
		sc, ok := cfg.Services[types.ServiceTypeHysteria2].(*hysteria2.ServerConfig)
		if !ok {
			return nil, fmt.Errorf("missing or wrong config type for service %q", t)
		}

		return hysteria2.NewServer("hysteria2", homeDir, sc), nil
	case types.ServiceTypeXray:
		sc, ok := cfg.Services[types.ServiceTypeXray].(*xray.ServerConfig)
		if !ok {
			return nil, fmt.Errorf("missing or wrong config type for service %q", t)
		}

		return xray.NewServer("xray", homeDir, sc), nil
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

			// Initialize all service configs if "skipService" is disabled.
			if !skipService {
				for _, t := range []types.ServiceType{
					types.ServiceTypeAmneziaWG,
					types.ServiceTypeHysteria2,
					types.ServiceTypeOpenVPN,
					types.ServiceTypeV2Ray,
					types.ServiceTypeWireGuard,
					types.ServiceTypeXray,
				} {
					log.Info("Initializing service", "type", t, "force", force)

					service, err := buildServerForInit(t, homeDir, cfg)
					if err != nil {
						return fmt.Errorf("building service %q: %w", t, err)
					}

					if err := service.Init(force); err != nil {
						return fmt.Errorf("running service %q init: %w", t, err)
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
