package core

import (
	"context"
	"fmt"

	"github.com/sentinel-official/sentinel-go-sdk/v2/libs/log"
)

// RemovePeerIfExists checks if a peer exists, and removes it if found.
func (c *Context) RemovePeerIfExists(ctx context.Context, id string) error {
	// Check if the peer exists.
	exists, err := c.Service().HasPeer(ctx, id)
	if err != nil {
		return fmt.Errorf("checking if peer %q exists in service: %w", id, err)
	}

	if !exists {
		return nil
	}

	// Remove the peer if it exists.
	if err := c.Service().RemovePeer(ctx, id); err != nil {
		return fmt.Errorf("removing peer %q from service: %w", id, err)
	}

	log.Info("Peer has been removed from service", "peer_id", id)

	return nil
}
