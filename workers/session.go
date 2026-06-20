package workers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/sentinel-official/sentinel-go-sdk/libs/cron"
	logger "github.com/sentinel-official/sentinel-go-sdk/libs/log"
	"github.com/sentinel-official/sentinelhub/v12/types/v1"
	"golang.org/x/sync/errgroup"

	"github.com/sentinel-official/sentinel-dvpnx/core"
	"github.com/sentinel-official/sentinel-dvpnx/database/models"
	"github.com/sentinel-official/sentinel-dvpnx/database/operations"
)

const (
	NameSessionUsageSyncWithBlockchain = "session_usage_sync_with_blockchain"
	NameSessionUsageSyncWithDatabase   = "session_usage_sync_with_database"
	NameSessionUsageValidate           = "session_usage_validate"
	NameSessionValidate                = "session_validate"
)

// aggregateSessionUsage sums rx/tx bytes and takes MAX duration across peers.
// Duration is MAX not SUM because summing would double-count overlapping protocols.
func aggregateSessionUsage(peers []models.SessionPeer) (rx, tx math.Int, duration time.Duration) {
	rx = math.ZeroInt()
	tx = math.ZeroInt()

	for i := range peers {
		rx = rx.Add(peers[i].GetRxBytes())
		tx = tx.Add(peers[i].GetTxBytes())

		if d := peers[i].GetDuration(); d > duration {
			duration = d
		}
	}

	return rx, tx, duration
}

// NewSessionUsageSyncWithBlockchainWorker creates a worker that aggregates each
// session's child peers and broadcasts usage updates to the blockchain.
func NewSessionUsageSyncWithBlockchainWorker(c *core.Context, interval time.Duration) cron.Worker {
	log := logger.With("module", "workers", "name", NameSessionUsageSyncWithBlockchain)

	handlerFunc := func(ctx context.Context) error {
		// Retrieve session records from the database.
		query := map[string]any{
			"node_addr": c.NodeAddr().String(),
		}

		items, err := operations.SessionFind(c.Database(), query)
		if err != nil {
			return fmt.Errorf("retrieving sessions from database: %w", err)
		}

		// Prepare a slice to collect messages.
		var (
			msgs []types.Msg
			mu   sync.Mutex
		)

		jobGroup, jobCtx := errgroup.WithContext(ctx)
		jobGroup.SetLimit(2)

		// Iterate over sessions and prepare messages for updates.
		for _, item := range items {
			jobGroup.Go(func() error {
				select {
				case <-jobCtx.Done():
					return nil
				default:
				}

				// Load the session's child peers and aggregate their usage.
				query := map[string]any{
					"session_id": item.GetID(),
				}

				peers, err := operations.SessionPeerFind(c.Database(), query)
				if err != nil {
					return fmt.Errorf("retrieving session_peers for session %d: %w", item.GetID(), err)
				}

				// Skip sessions without any peers.
				if len(peers) == 0 {
					log.Debug("Skipping session", "id", item.GetID(), "cause", "no peers")

					return nil
				}

				rx, tx, duration := aggregateSessionUsage(peers)

				session, err := c.Client().Session(jobCtx, item.GetID())
				if err != nil {
					return fmt.Errorf("querying session %d from blockchain: %w", item.GetID(), err)
				}

				// Skip session if it is nil
				if session == nil {
					log.Debug("Skipping session", "id", item.GetID(), "cause", "nil session")

					return nil
				}

				// Node Tx = client download; node Rx = client upload.
				downloadBytes, uploadBytes := tx, rx

				// Skip session if it is already up-to-date (rx maps to upload bytes on chain).
				if session.GetUploadBytes().Equal(uploadBytes) {
					log.Debug("Skipping session", "id", item.GetID(), "cause", "already up-to-date")

					return nil
				}

				// Generate an update message for the session from the aggregated usage.
				msg := item.MsgUpdateSessionRequest(downloadBytes, uploadBytes, duration)
				log.Debug("Adding session to update list",
					"id", item.GetID(), "download_bytes", msg.DownloadBytes,
					"duration", msg.Duration, "upload_bytes", msg.UploadBytes,
				)

				mu.Lock()
				defer mu.Unlock()

				msgs = append(msgs, msg)

				return nil
			})
		}

		// Wait until all routines complete.
		if err := jobGroup.Wait(); err != nil {
			return fmt.Errorf("waiting job group: %w", err)
		}

		// Broadcast the prepared messages as a transaction.
		if err := c.BroadcastTx(ctx, msgs...); err != nil {
			return fmt.Errorf("broadcasting tx with %d update_session msg(s): %w", len(msgs), err)
		}

		return nil
	}

	// Initialize and return the worker.
	return cron.NewBasicWorker(NameSessionUsageSyncWithBlockchain).
		WithHandler(handlerFunc).
		WithInterval(interval).
		WithRetryDelay(5 * time.Second)
}

// NewSessionUsageSyncWithDatabaseWorker updates session usage in the database per service.
// Per-service maps are never merged: WG and AWG share key space so merging would cross-attribute usage.
func NewSessionUsageSyncWithDatabaseWorker(c *core.Context, interval time.Duration) cron.Worker {
	log := logger.With("module", "workers", "name", NameSessionUsageSyncWithDatabase)

	handlerFunc := func(ctx context.Context) error {
		jobGroup, jobCtx := errgroup.WithContext(ctx)
		jobGroup.SetLimit(2)

		// Fan out over each active service; PeerStatistics() is fetched inside the closure
		// so errors propagate through the group without abandoning sibling goroutines.
		for serviceType, service := range c.Services() {
			jobGroup.Go(func() error {
				select {
				case <-jobCtx.Done():
					return nil
				default:
				}

				// Fetch peer usage statistics from this service.
				stats, err := service.PeerStatistics()
				if err != nil {
					return fmt.Errorf("retrieving peer statistics from service %q: %w", serviceType, err)
				}

				// Update the database with the fetched statistics for this service only.
				for peerID, item := range stats {
					select {
					case <-jobCtx.Done():
						return nil
					default:
					}

					if time.Since(item.UpdatedAt) > interval {
						log.Debug("Skipping peer",
							"service_type", serviceType, "peer_id", peerID, "cause", "already up-to-date",
							"updated_at", item.UpdatedAt,
						)

						continue
					}

					// Convert usage statistics to strings for database storage.
					rxBytes := math.NewInt(item.RxBytes).String()
					txBytes := math.NewInt(item.TxBytes).String()

					// Locate the session_peer by (service_type, peer_id).
					query := map[string]any{
						"service_type": serviceType.String(),
						"peer_id":      peerID,
					}

					// Define updates to apply to the session_peer record.
					updates := map[string]any{
						"rx_bytes": rxBytes,
						"tx_bytes": txBytes,
					}

					log.Debug("Updating session_peer in database",
						"service_type", serviceType, "peer_id", peerID,
						"rx_bytes", rxBytes, "tx_bytes", txBytes,
					)

					if _, err := operations.SessionPeerFindOneAndUpdate(c.Database(), query, updates); err != nil {
						return fmt.Errorf("updating session_peer %q from service %q: %w", peerID, serviceType, err)
					}
				}

				return nil
			})
		}

		// Wait until all routines complete.
		if err := jobGroup.Wait(); err != nil {
			return fmt.Errorf("waiting job group: %w", err)
		}

		return nil
	}

	// Initialize and return the worker.
	return cron.NewBasicWorker(NameSessionUsageSyncWithDatabase).
		WithHandler(handlerFunc).
		WithInterval(interval)
}

// NewSessionUsageValidateWorker validates per-session budgets shared across protocols:
// max_bytes vs SUM(rx+tx) and max_duration vs MAX(duration); on exceed, all peers are removed.
func NewSessionUsageValidateWorker(c *core.Context, interval time.Duration) cron.Worker {
	log := logger.With("module", "workers", "name", NameSessionUsageValidate)

	handlerFunc := func(ctx context.Context) error {
		// Retrieve session records from the database.
		query := map[string]any{
			"node_addr": c.NodeAddr().String(),
		}

		items, err := operations.SessionFind(c.Database(), query)
		if err != nil {
			return fmt.Errorf("retrieving sessions from database: %w", err)
		}

		jobGroup, jobCtx := errgroup.WithContext(ctx)
		jobGroup.SetLimit(2)

		// Validate session limits and remove peers if needed.
		for _, item := range items {
			jobGroup.Go(func() error {
				select {
				case <-jobCtx.Done():
					return nil
				default:
				}

				// Load the session's child peers and aggregate their usage.
				query := map[string]any{
					"session_id": item.GetID(),
				}

				peers, err := operations.SessionPeerFind(c.Database(), query)
				if err != nil {
					return fmt.Errorf("retrieving session_peers for session %d: %w", item.GetID(), err)
				}

				if len(peers) == 0 {
					return nil
				}

				rx, tx, duration := aggregateSessionUsage(peers)
				totalBytes := rx.Add(tx)

				removePeers := false

				// Check if the session exceeds the maximum allowed bytes (SUM(rx+tx)).
				maxBytes := item.GetMaxBytes()
				if !maxBytes.IsZero() && totalBytes.GTE(maxBytes) {
					log.Debug("Marking session for peer removal",
						"id", item.GetID(), "cause", "exceeds max bytes",
						"total_bytes", totalBytes, "max_bytes", maxBytes,
					)

					removePeers = true
				}

				// Check if the session exceeds the maximum allowed duration (MAX(duration)).
				maxDuration := item.GetMaxDuration()
				if maxDuration != 0 && duration >= maxDuration {
					log.Debug("Marking session for peer removal",
						"id", item.GetID(), "cause", "exceeds max duration",
						"duration", duration, "max_duration", maxDuration,
					)

					removePeers = true
				}

				// If the session exceeded any limits, remove every peer across services.
				if removePeers {
					for i := range peers {
						log.Debug("Removing peer from service",
							"id", item.GetID(), "service_type", peers[i].GetServiceType(),
							"peer_id", peers[i].GetPeerID(),
						)

						if err := c.RemovePeerIfExists(jobCtx, peers[i].GetServiceType(), peers[i].GetPeerID()); err != nil {
							return fmt.Errorf("removing peer %q from service %q for session %d: %w",
								peers[i].GetPeerID(), peers[i].GetServiceType(), item.GetID(), err)
						}
					}
				}

				return nil
			})
		}

		// Wait until all routines complete.
		if err := jobGroup.Wait(); err != nil {
			return fmt.Errorf("waiting job group: %w", err)
		}

		return nil
	}

	// Initialize and return the worker.
	return cron.NewBasicWorker(NameSessionUsageValidate).
		WithHandler(handlerFunc).
		WithInterval(interval)
}

// NewSessionValidateWorker validates session status against the chain; when a session is
// missing or inactive, all child peers are removed and the session and its peers deleted.
func NewSessionValidateWorker(c *core.Context, interval time.Duration) cron.Worker {
	log := logger.With("module", "workers", "name", NameSessionValidate)

	handlerFunc := func(ctx context.Context) error {
		// Retrieve session records from the database.
		query := map[string]any{
			"node_addr": c.NodeAddr().String(),
		}

		items, err := operations.SessionFind(c.Database(), query)
		if err != nil {
			return fmt.Errorf("retrieving sessions from database: %w", err)
		}

		jobGroup, jobCtx := errgroup.WithContext(ctx)
		jobGroup.SetLimit(2)

		// Validate session status and consistency.
		for _, item := range items {
			jobGroup.Go(func() error {
				select {
				case <-jobCtx.Done():
					return nil
				default:
				}

				session, err := c.Client().Session(jobCtx, item.GetID())
				if err != nil {
					return fmt.Errorf("querying session %d from blockchain: %w", item.GetID(), err)
				}

				// Determine whether the session must be torn down.
				remove := false

				if session == nil {
					log.Debug("Marking session for removal",
						"id", item.GetID(), "cause", "nil session",
					)

					remove = true
				}

				if session != nil && !session.GetStatus().Equal(v1.StatusActive) {
					log.Debug("Marking session for removal",
						"id", item.GetID(), "cause", "invalid session status",
						"got", session.GetStatus(), "expected", v1.StatusActive,
					)

					remove = true
				}

				if !remove {
					return nil
				}

				// Load the session's child peers.
				query := map[string]any{
					"session_id": item.GetID(),
				}

				peers, err := operations.SessionPeerFind(c.Database(), query)
				if err != nil {
					return fmt.Errorf("retrieving session_peers for session %d: %w", item.GetID(), err)
				}

				// Remove every peer (routed by its service_type) and delete the child row.
				for i := range peers {
					log.Debug("Removing peer from service",
						"id", item.GetID(), "service_type", peers[i].GetServiceType(),
						"peer_id", peers[i].GetPeerID(),
					)

					if err := c.RemovePeerIfExists(jobCtx, peers[i].GetServiceType(), peers[i].GetPeerID()); err != nil {
						return fmt.Errorf("removing peer %q from service %q for session %d: %w",
							peers[i].GetPeerID(), peers[i].GetServiceType(), item.GetID(), err)
					}

					// Delete the child row explicitly (belt-and-suspenders alongside the FK cascade).
					peerQuery := map[string]any{
						"session_id":   item.GetID(),
						"service_type": peers[i].GetServiceType().String(),
					}

					if _, err := operations.SessionPeerFindOneAndDelete(c.Database(), peerQuery); err != nil {
						return fmt.Errorf("deleting session_peer %q from service %q for session %d: %w",
							peers[i].GetPeerID(), peers[i].GetServiceType(), item.GetID(), err)
					}
				}

				// Delete the parent session record (children cascade via FK).
				log.Info("Deleting session from database", "id", item.GetID())

				sessionQuery := map[string]any{
					"id": item.GetID(),
				}

				if _, err := operations.SessionFindOneAndDelete(c.Database(), sessionQuery); err != nil {
					return fmt.Errorf("deleting session %d from database: %w", item.GetID(), err)
				}

				return nil
			})
		}

		// Wait until all routines complete.
		if err := jobGroup.Wait(); err != nil {
			return fmt.Errorf("waiting job group: %w", err)
		}

		return nil
	}

	// Initialize and return the worker.
	return cron.NewBasicWorker(NameSessionValidate).
		WithHandler(handlerFunc).
		WithInterval(interval)
}
