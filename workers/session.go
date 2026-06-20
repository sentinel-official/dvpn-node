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

// aggregateSessionUsage aggregates a session's child-peer usage: it sums rx and
// tx across the peers and takes the MAX duration. Summing duration would
// double-count overlapping protocols and a single peer would miss protocols
// active at different times, so the session duration is the maximum elapsed
// across its peers. An empty slice yields zero values.
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

// NewSessionUsageSyncWithBlockchainWorker creates a worker that synchronizes session usage with the blockchain.
// This worker retrieves session data from the database, aggregates each session's child peers,
// and broadcasts any updates as transactions.
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
		for _, val := range items {
			item := val

			jobGroup.Go(func() error {
				select {
				case <-jobCtx.Done():
					return nil
				default:
				}

				// Load the session's child peers and aggregate their usage.
				peers, err := operations.SessionPeerFind(c.Database(), map[string]any{
					"session_id": item.GetID(),
				})
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

				// Skip session if it is already up-to-date (rx maps to upload bytes on chain).
				if session.GetUploadBytes().Equal(rx) {
					log.Debug("Skipping session", "id", item.GetID(), "cause", "already up-to-date")
					return nil
				}

				// Generate an update message for the session from the aggregated usage.
				msg := item.MsgUpdateSessionRequest(tx, rx, duration)
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

// NewSessionUsageSyncWithDatabaseWorker creates a worker that updates session usage in the database.
// This worker fetches usage data from each active service and updates the matching SessionPeer
// record by (service_type, peer_id). The per-service maps are never merged: WG and AWG peer IDs
// are the client public key, so the same key can appear under both protocols and merging would
// cross-attribute usage.
func NewSessionUsageSyncWithDatabaseWorker(c *core.Context, interval time.Duration) cron.Worker {
	log := logger.With("module", "workers", "name", NameSessionUsageSyncWithDatabase)

	handlerFunc := func(ctx context.Context) error {
		jobGroup, jobCtx := errgroup.WithContext(ctx)
		jobGroup.SetLimit(2)

		// Fan out over each active service; process each service's statistics under its own type.
		for serviceType, svc := range c.Services() {
			t, service := serviceType, svc

			// Fetch peer usage statistics from this service.
			stats, err := service.PeerStatistics()
			if err != nil {
				return fmt.Errorf("retrieving peer statistics from service %q: %w", t, err)
			}

			// Update the database with the fetched statistics for this service only.
			for key, val := range stats {
				peerID, item := key, val

				jobGroup.Go(func() error {
					select {
					case <-jobCtx.Done():
						return nil
					default:
					}

					if time.Since(item.UpdatedAt) > interval {
						log.Debug("Skipping peer",
							"service_type", t, "peer_id", peerID, "cause", "already up-to-date",
							"updated_at", item.UpdatedAt,
						)

						return nil
					}

					// Convert usage statistics to strings for database storage.
					rxBytes := math.NewInt(item.RxBytes).String()
					txBytes := math.NewInt(item.TxBytes).String()

					// Locate the session_peer by (service_type, peer_id).
					query := map[string]any{
						"service_type": t.String(),
						"peer_id":      peerID,
					}

					// Define updates to apply to the session_peer record.
					updates := map[string]any{
						"rx_bytes": rxBytes,
						"tx_bytes": txBytes,
					}

					log.Debug("Updating session_peer in database",
						"service_type", t, "peer_id", peerID, "rx_bytes", rxBytes, "tx_bytes", txBytes,
					)

					if _, err := operations.SessionPeerFindOneAndUpdate(c.Database(), query, updates); err != nil {
						return fmt.Errorf("updating session_peer for service %q peer %q: %w", t, peerID, err)
					}

					return nil
				})
			}
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

// NewSessionUsageValidateWorker creates a worker that validates session usage limits and removes peers if necessary.
// Budgets (max_bytes/max_duration) are per-session and shared across the session's protocols: max_bytes is checked
// against SUM(rx+tx) over the child peers and max_duration against MAX(duration). On exceed, all peers of the
// session are removed across services.
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
		for _, val := range items {
			item := val

			jobGroup.Go(func() error {
				select {
				case <-jobCtx.Done():
					return nil
				default:
				}

				// Load the session's child peers and aggregate their usage.
				peers, err := operations.SessionPeerFind(c.Database(), map[string]any{
					"session_id": item.GetID(),
				})
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
						p := peers[i]

						log.Debug("Removing peer from service",
							"id", item.GetID(), "service_type", p.GetServiceType(), "peer_id", p.GetPeerID(),
						)

						if err := c.RemovePeerIfExists(jobCtx, p.GetServiceType(), p.GetPeerID()); err != nil {
							return fmt.Errorf("removing peer %q (%s) for session %d from service: %w",
								p.GetPeerID(), p.GetServiceType(), item.GetID(), err)
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

// NewSessionValidateWorker creates a worker that validates session status and removes peers if necessary.
// This worker ensures sessions are active and consistent between the database and blockchain. When a session
// is missing or inactive on chain, every child peer is removed (routed by its service_type) and the session
// row plus its session_peers are deleted.
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
		for _, val := range items {
			item := val

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
				peers, err := operations.SessionPeerFind(c.Database(), map[string]any{
					"session_id": item.GetID(),
				})
				if err != nil {
					return fmt.Errorf("retrieving session_peers for session %d: %w", item.GetID(), err)
				}

				// Remove every peer (routed by its service_type) and delete the child row.
				for i := range peers {
					p := peers[i]

					log.Debug("Removing peer from service",
						"id", item.GetID(), "service_type", p.GetServiceType(), "peer_id", p.GetPeerID(),
					)

					if err := c.RemovePeerIfExists(jobCtx, p.GetServiceType(), p.GetPeerID()); err != nil {
						return fmt.Errorf("removing peer %q (%s) for session %d from service: %w",
							p.GetPeerID(), p.GetServiceType(), item.GetID(), err)
					}

					// Delete the child row explicitly (belt-and-suspenders alongside the FK cascade).
					if _, err := operations.SessionPeerFindOneAndDelete(c.Database(), map[string]any{
						"session_id":   item.GetID(),
						"service_type": p.GetServiceType().String(),
					}); err != nil {
						return fmt.Errorf("deleting session_peer %q (%s) for session %d: %w",
							p.GetPeerID(), p.GetServiceType(), item.GetID(), err)
					}
				}

				// Delete the parent session record (children cascade via FK).
				log.Info("Deleting session from database", "id", item.GetID())

				if _, err := operations.SessionFindOneAndDelete(c.Database(), map[string]any{
					"id": item.GetID(),
				}); err != nil {
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
