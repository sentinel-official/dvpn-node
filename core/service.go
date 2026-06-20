package core

import (
	"context"
	"fmt"

	"github.com/sentinel-official/sentinel-go-sdk/libs/log"
	sentinelsdk "github.com/sentinel-official/sentinel-go-sdk/types"
)

// RemovePeerIfExists checks if a peer exists for the given service type, and removes it if found.
// If the service type is not active in the registry, the peer cannot be running, so removal is
// vacuously satisfied and the function returns nil (allowing callers to reap stale DB rows).
func (c *Context) RemovePeerIfExists(ctx context.Context, t sentinelsdk.ServiceType, peerID string) error {
	// Resolve the service for the given type.
	svc, ok := c.ServiceFor(t)
	if !ok {
		log.Warn("Peer's service type is not active; nothing to remove", "peer_id", peerID, "service_type", t)
		return nil
	}

	// Check if the peer exists.
	exists, err := svc.HasPeer(ctx, peerID)
	if err != nil {
		return fmt.Errorf("checking if peer %q exists in service: %w", peerID, err)
	}

	if !exists {
		return nil
	}

	// Remove the peer if it exists.
	if err := svc.RemovePeer(ctx, peerID); err != nil {
		return fmt.Errorf("removing peer %q from service: %w", peerID, err)
	}

	log.Info("Peer has been removed from service", "peer_id", peerID, "service_type", t)

	return nil
}
