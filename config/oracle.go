package config

import (
	"errors"
	"fmt"

	"github.com/spf13/pflag"
)

// CoinGeckoConfig holds settings for the CoinGecko oracle.
type CoinGeckoConfig struct {
	APIKey string `mapstructure:"api_key"` // APIKey specifies the API key for CoinGecko.
}

// GetAPIKey returns the APIKey field.
func (c *CoinGeckoConfig) GetAPIKey() string {
	return c.APIKey
}

// Validate checks the validity of the CoinGeckoConfig configuration.
func (c *CoinGeckoConfig) Validate() error {
	return nil
}

// DefaultCoinGeckoConfig returns an CoinGeckoConfig instance with default values.
func DefaultCoinGeckoConfig() *CoinGeckoConfig {
	return &CoinGeckoConfig{
		APIKey: "",
	}
}

// OsmosisConfig holds settings for the Osmosis oracle.
type OsmosisConfig struct {
	APIAddr string `mapstructure:"api_addr"` // APIAddr specifies the Osmosis API endpoint.
}

// GetAPIAddr returns the APIAddr field.
func (c *OsmosisConfig) GetAPIAddr() string {
	return c.APIAddr
}

// Validate checks the validity of the OsmosisConfig configuration.
func (c *OsmosisConfig) Validate() error {
	if c.APIAddr == "" {
		return errors.New("api_addr cannot be empty")
	}

	return nil
}

// DefaultOsmosisConfig returns an OsmosisConfig instance with default values.
func DefaultOsmosisConfig() *OsmosisConfig {
	return &OsmosisConfig{
		APIAddr: "https://lcd.osmosis.zone:443",
	}
}

// OracleConfig represents the configuration for oracles such as Osmosis and CoinGecko.
type OracleConfig struct {
	Name      string           `mapstructure:"name"`      // Name specifies the oracle's name.
	CoinGecko *CoinGeckoConfig `mapstructure:"coingecko"` // CoinGecko configuration.
	Osmosis   *OsmosisConfig   `mapstructure:"osmosis"`   // Osmosis configuration.
}

// WithName sets the Name field and returns the updated OracleConfig.
func (c *OracleConfig) WithName(name string) *OracleConfig {
	c.Name = name

	return c
}

// GetName returns the Name field.
func (c *OracleConfig) GetName() string {
	return c.Name
}

// Validate checks the validity of the OracleConfig configuration.
func (c *OracleConfig) Validate() error {
	if c.Name == "" {
		return nil
	}

	validNames := map[string]bool{
		"coingecko": true,
		"osmosis":   true,
	}

	if !validNames[c.Name] {
		return fmt.Errorf("unsupported name %q (allowed: coingecko, osmosis)", c.Name)
	}

	switch c.Name {
	case "coingecko":
		if c.CoinGecko == nil {
			return errors.New("coingecko config cannot be nil")
		}

		if err := c.CoinGecko.Validate(); err != nil {
			return fmt.Errorf("validating coingecko config: %w", err)
		}

	case "osmosis":
		if c.Osmosis == nil {
			return errors.New("osmosis config cannot be nil")
		}

		if err := c.Osmosis.Validate(); err != nil {
			return fmt.Errorf("validating osmosis config: %w", err)
		}

	default:
		return fmt.Errorf("unsupported name %q", c.Name)
	}

	return nil
}

// SetForFlags adds oracle configuration flags to the specified FlagSet.
func (c *OracleConfig) SetForFlags(f *pflag.FlagSet) {
	f.StringVar(&c.Name, "oracle.name", c.Name, "specify which oracle provider to use (e.g., coingecko or osmosis)")
	f.StringVar(&c.CoinGecko.APIKey, "oracle.coingecko.api-key", c.CoinGecko.APIKey, "set the API key used to authenticate requests to the CoinGecko oracle")
	f.StringVar(&c.Osmosis.APIAddr, "oracle.osmosis.api-addr", c.Osmosis.APIAddr, "set the API endpoint for the Osmosis oracle")
}

// DefaultOracleConfig returns an OracleConfig instance with default values.
func DefaultOracleConfig() *OracleConfig {
	return &OracleConfig{
		Name:      "coingecko",
		CoinGecko: DefaultCoinGeckoConfig(),
		Osmosis:   DefaultOsmosisConfig(),
	}
}
