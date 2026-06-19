package config

import (
	"errors"
	"fmt"

	"github.com/spf13/pflag"
)

const MaxHandshakeDNSPeers = 1 << 3 // Maximum number of peers for Handshake DNS.

const DefaultHandshakeDNSMaxRestarts = 5 // Default hnsd restart budget before the node is stopped.

// HandshakeDNSConfig represents the Handshake DNS configuration.
type HandshakeDNSConfig struct {
	Enable      bool `mapstructure:"enable"`       // Enable specifies if Handshake DNS is enabled.
	Peers       uint `mapstructure:"peers"`        // Peers specifies the number of DNS peers.
	MaxRestarts int  `mapstructure:"max_restarts"` // MaxRestarts bounds hnsd restarts (0 disables, -1 unlimited).
}

// WithEnable sets the Enable field and returns the updated HandshakeDNSConfig.
func (c *HandshakeDNSConfig) WithEnable(enable bool) *HandshakeDNSConfig {
	c.Enable = enable

	return c
}

// WithPeers sets the Peers field and returns the updated HandshakeDNSConfig.
func (c *HandshakeDNSConfig) WithPeers(peers uint) *HandshakeDNSConfig {
	c.Peers = peers

	return c
}

// GetEnable returns the Enable field.
func (c *HandshakeDNSConfig) GetEnable() bool {
	return c.Enable
}

// GetPeers returns the Peers field.
func (c *HandshakeDNSConfig) GetPeers() uint {
	return c.Peers
}

// WithMaxRestarts sets the MaxRestarts field and returns the updated HandshakeDNSConfig.
func (c *HandshakeDNSConfig) WithMaxRestarts(maxRestarts int) *HandshakeDNSConfig {
	c.MaxRestarts = maxRestarts

	return c
}

// GetMaxRestarts returns the MaxRestarts field.
func (c *HandshakeDNSConfig) GetMaxRestarts() int {
	return c.MaxRestarts
}

// Validate checks the validity of the HandshakeDNSConfig configuration.
func (c *HandshakeDNSConfig) Validate() error {
	// If Handshake DNS is not enabled, validation passes.
	if !c.Enable {
		return nil
	}

	// Ensure the number of peers is not zero.
	if c.Peers == 0 {
		return errors.New("peers cannot be zero")
	}

	// Ensure the number of peers does not exceed the maximum allowed value.
	if c.Peers > MaxHandshakeDNSPeers {
		return fmt.Errorf("peers cannot be greater than %d", MaxHandshakeDNSPeers)
	}

	if c.MaxRestarts < -1 {
		return errors.New("max_restarts cannot be less than -1")
	}

	return nil
}

// SetForFlags adds handshake-dns configuration flags to the specified FlagSet.
func (c *HandshakeDNSConfig) SetForFlags(f *pflag.FlagSet) {
	f.BoolVar(&c.Enable, "handshake-dns.enable", c.Enable, "enable or disable Handshake DNS")
	f.UintVar(&c.Peers, "handshake-dns.peers", c.Peers, "number of Handshake DNS peers")
	f.IntVar(&c.MaxRestarts, "handshake-dns.max-restarts", c.MaxRestarts, "maximum hnsd restarts before failing the node (0 disables, -1 for unlimited)")
}

// DefaultHandshakeDNSConfig returns a HandshakeDNSConfig instance with default values.
func DefaultHandshakeDNSConfig() *HandshakeDNSConfig {
	return &HandshakeDNSConfig{
		Enable:      false,
		Peers:       MaxHandshakeDNSPeers,
		MaxRestarts: DefaultHandshakeDNSMaxRestarts,
	}
}
