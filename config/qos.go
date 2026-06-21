package config

import (
	"errors"
	"fmt"

	"github.com/spf13/pflag"
)

const MaxQoSMaxSessions = 250 // Maximum allowed value for MaxSessions.

// QoSConfig represents the Quality of Service (QoS) configuration.
type QoSConfig struct {
	MaxSessions uint `mapstructure:"max_sessions"` // MaxSessions specifies the maximum number of sessions.
}

// WithMaxSessions sets the MaxSessions field and returns the updated QoSConfig.
func (c *QoSConfig) WithMaxSessions(maxSessions uint) *QoSConfig {
	c.MaxSessions = maxSessions

	return c
}

// GetMaxSessions returns the MaxSessions field.
func (c *QoSConfig) GetMaxSessions() uint {
	return c.MaxSessions
}

// Validate checks the validity of the QoS configuration.
func (c *QoSConfig) Validate() error {
	// Ensure MaxSessions is not zero.
	if c.MaxSessions == 0 {
		return errors.New("max_sessions cannot be zero")
	}

	// Ensure MaxSessions does not exceed the maximum allowed value.
	if c.MaxSessions > MaxQoSMaxSessions {
		return fmt.Errorf("max_sessions cannot be greater than %d", MaxQoSMaxSessions)
	}

	return nil
}

// SetForFlags adds qos configuration flags to the specified FlagSet.
func (c *QoSConfig) SetForFlags(f *pflag.FlagSet) {
	f.UintVar(&c.MaxSessions, "qos.max-sessions", c.MaxSessions, "maximum number of sessions for service")
}

// DefaultQoSConfig returns a QoSConfig instance with default values.
func DefaultQoSConfig() *QoSConfig {
	return &QoSConfig{
		MaxSessions: MaxQoSMaxSessions,
	}
}
