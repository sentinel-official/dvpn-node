package config

import (
	"embed"
	"errors"
	"fmt"
	"net/netip"
	"os"

	"github.com/sentinel-official/sentinel-go-sdk/amneziawg"
	"github.com/sentinel-official/sentinel-go-sdk/core/config"
	"github.com/sentinel-official/sentinel-go-sdk/types"
	"github.com/sentinel-official/sentinel-go-sdk/utils"
	"github.com/sentinel-official/sentinel-go-sdk/wireguard"
	"github.com/spf13/pflag"
)

// Embed the template files for configuration.
//
//go:embed *.tmpl
var fs embed.FS

// Config represents the overall configuration structure.
type Config struct {
	*config.Config `mapstructure:",squash"`

	HandshakeDNS *HandshakeDNSConfig `mapstructure:"handshake_dns"` // HandshakeDNS contains Handshake DNS configuration.
	Node         *NodeConfig         `mapstructure:"node"`          // Node contains node-specific configuration.
	Oracle       *OracleConfig       `mapstructure:"oracle"`        // Oracle contains oracle-specific configuration.
	QoS          *QoSConfig          `mapstructure:"qos"`           // QoS contains Quality of Service configuration.

	Services map[types.ServiceType]types.ServiceConfig `mapstructure:"-"`
}

// Validate validates the entire configuration.
func (c *Config) Validate() error {
	if err := c.Config.Validate(); err != nil {
		return fmt.Errorf("validating base config: %w", err)
	}

	if err := c.HandshakeDNS.Validate(); err != nil {
		return fmt.Errorf("validating handshake_dns config: %w", err)
	}

	if err := c.Node.Validate(); err != nil {
		return fmt.Errorf("validating node config: %w", err)
	}

	if err := c.Oracle.Validate(); err != nil {
		return fmt.Errorf("validating oracle config: %w", err)
	}

	if err := c.QoS.Validate(); err != nil {
		return fmt.Errorf("validating QoS config: %w", err)
	}

	if c.HandshakeDNS.GetEnable() {
		hasWireGuard := false

		for _, t := range c.Node.GetServiceTypes() {
			if t == types.ServiceTypeWireGuard || t == types.ServiceTypeAmneziaWG {
				hasWireGuard = true

				break
			}
		}

		if !hasWireGuard {
			return errors.New("handshake_dns requires at least one of wireguard or amneziawg in service_types")
		}
	}

	return nil
}

// ServiceGatewayAddr returns the WireGuard/AmneziaWG tunnel gateway IP,
// preferring WireGuard when both are enabled.
func (c *Config) ServiceGatewayAddr() (string, error) {
	enabled := make(map[types.ServiceType]bool)
	for _, t := range c.Node.GetServiceTypes() {
		enabled[t] = true
	}

	var ipv4Addr, ipv6Addr string

	switch {
	case enabled[types.ServiceTypeWireGuard]:
		v := c.Services[types.ServiceTypeWireGuard].(*wireguard.ServerConfig)
		ipv4Addr, ipv6Addr = v.IPv4Addr, v.IPv6Addr
	case enabled[types.ServiceTypeAmneziaWG]:
		v := c.Services[types.ServiceTypeAmneziaWG].(*amneziawg.ServerConfig)
		ipv4Addr, ipv6Addr = v.IPv4Addr, v.IPv6Addr
	default:
		return "", errors.New("handshake_dns requires wireguard or amneziawg in service_types")
	}

	addr := ipv4Addr
	if addr == "" {
		addr = ipv6Addr
	}

	if addr == "" {
		return "", errors.New("service has no ipv4 or ipv6 addr configured")
	}

	prefix, err := netip.ParsePrefix(addr)
	if err != nil {
		return "", fmt.Errorf("parsing addr %q: %w", addr, err)
	}

	return prefix.Addr().String(), nil
}

// SetForFlags adds configuration flags to the specified FlagSet.
func (c *Config) SetForFlags(f *pflag.FlagSet) {
	c.Config.SetForFlags(f)
	c.HandshakeDNS.SetForFlags(f)
	c.Node.SetForFlags(f)
	c.Oracle.SetForFlags(f)
	c.QoS.SetForFlags(f)
}

// DefaultConfig returns a configuration instance with default values.
func DefaultConfig() *Config {
	return &Config{
		Config:       config.DefaultConfig(),
		HandshakeDNS: DefaultHandshakeDNSConfig(),
		Node:         DefaultNodeConfig(),
		Oracle:       DefaultOracleConfig(),
		QoS:          DefaultQoSConfig(),
	}
}

// WriteAppConfig generates the application-level configuration file using the main config template.
func (c *Config) WriteAppConfig(file string) error {
	// Load the application template from the embedded filesystem.
	text, err := fs.ReadFile("config.toml.tmpl")
	if err != nil {
		return fmt.Errorf("reading config template: %w", err)
	}

	// Render the template with Config data and write the result to the specified file.
	if err := utils.ExecTemplateToFile(string(text), c, file); err != nil {
		return fmt.Errorf("writing rendered config file %q: %w", file, err)
	}

	// Restrict file permissions to owner read/write only.
	if err := os.Chmod(file, 0600); err != nil {
		return fmt.Errorf("setting file permissions: %w", err)
	}

	return nil
}
