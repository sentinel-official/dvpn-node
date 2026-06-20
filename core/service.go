package core

import (
	"context"
	"fmt"

	"github.com/sentinel-official/sentinel-go-sdk/libs/log"
	sentinelsdk "github.com/sentinel-official/sentinel-go-sdk/types"
)

// RemovePeerIfExists removes the peer for the given service type if present.
// An inactive service type means the peer cannot be running, so it returns nil.
func (c *Context) RemovePeerIfExists(ctx context.Context, t sentinelsdk.ServiceType, peerID string) error {
	// Resolve the service for the given type.
	service, ok := c.Service(t)
	if !ok {
		return nil
	}

	// Check if the peer exists.
	exists, err := service.HasPeer(ctx, peerID)
	if err != nil {
		return fmt.Errorf("checking if peer %q exists in service: %w", peerID, err)
	}

	if !exists {
		return nil
	}

	// Remove the peer if it exists.
	if err := service.RemovePeer(ctx, peerID); err != nil {
		return fmt.Errorf("removing peer %q from service: %w", peerID, err)
	}

	log.Info("Peer has been removed from service", "peer_id", peerID, "service_type", t)

	return nil
}
