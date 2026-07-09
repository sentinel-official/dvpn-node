package workers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/sentinel-official/sentinel-go-sdk/v2/libs/cron"
	logger "github.com/sentinel-official/sentinel-go-sdk/v2/libs/log"
	"github.com/sentinel-official/sentinelhub/v12/types/v1"
	"golang.org/x/sync/errgroup"

	"github.com/sentinel-official/sentinel-dvpnx/core"
	"github.com/sentinel-official/sentinel-dvpnx/database/operations"
)

const (
	NameSessionUsageSyncWithBlockchain = "session_usage_sync_with_blockchain"
	NameSessionUsageSyncWithDatabase   = "session_usage_sync_with_database"
	NameSessionUsageValidate           = "session_usage_validate"
	NameSessionValidate                = "session_validate"
)

// NewSessionUsageSyncWithBlockchainWorker creates a worker that synchronizes session usage with the blockchain.
// This worker retrieves session data from the database, validates it against the blockchain,
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

				session, err := c.Client().Session(jobCtx, item.GetID())
				if err != nil {
					return fmt.Errorf("querying session %d from blockchain: %w", item.GetID(), err)
				}

				// Skip session if it is nil
				if session == nil {
					log.Debug("Skipping session",
						"id", item.GetID(), "peer_id", item.GetPeerID(), "cause", "nil session",
					)

					return nil
				}

				// Skip session if it is already up-to-date
				if session.GetUploadBytes().Equal(item.GetRxBytes()) {
					log.Debug("Skipping session",
						"id", item.GetID(), "peer_id", item.GetPeerID(), "cause", "already up-to-date",
					)

					return nil
				}

				// Generate an update message for the session.
				msg := item.MsgUpdateSessionRequest()
				log.Debug("Adding session to update list",
					"id", item.GetID(), "peer_id", item.GetPeerID(), "download_bytes", msg.DownloadBytes,
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
// This worker fetches usage data from the peer service and updates the corresponding database records.
func NewSessionUsageSyncWithDatabaseWorker(c *core.Context, interval time.Duration) cron.Worker {
	log := logger.With("module", "workers", "name", NameSessionUsageSyncWithDatabase)

	handlerFunc := func(ctx context.Context) error {
		// Fetch peer usage statistics from the service.
		items, err := c.Service().PeerStatistics()
		if err != nil {
			return fmt.Errorf("retrieving peer statistics from service: %w", err)
		}

		jobGroup, jobCtx := errgroup.WithContext(ctx)
		jobGroup.SetLimit(2)

		// Update the database with the fetched statistics.
		for key, val := range items {
			peerID, item := key, val

			jobGroup.Go(func() error {
				select {
				case <-jobCtx.Done():
					return nil
				default:
				}

				if time.Since(item.UpdatedAt) > interval {
					log.Debug("Skipping session",
						"id", 0, "peer_id", peerID, "cause", "already up-to-date",
						"updated_at", item.UpdatedAt,
					)

					return nil
				}

				// Convert usage statistics to strings for database storage.
				rxBytes := math.NewInt(item.RxBytes).String()
				txBytes := math.NewInt(item.TxBytes).String()

				// Define query to find the session by peer id.
				query := map[string]any{
					"peer_id": peerID,
				}

				// Define updates to apply to the session record.
				updates := map[string]any{
					"rx_bytes": rxBytes,
					"tx_bytes": txBytes,
				}

				log.Debug("Updating session in database",
					"id", 0, "peer_id", peerID, "rx_bytes", rxBytes, "tx_bytes", txBytes,
				)

				if _, err := operations.SessionFindOneAndUpdate(c.Database(), query, updates); err != nil {
					return fmt.Errorf("updating session for peer %q in database: %w", peerID, err)
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

// NewSessionUsageValidateWorker creates a worker that validates session usage limits and removes peers if necessary.
// This worker checks if sessions exceed their maximum byte or duration limits and removes peers accordingly.
func NewSessionUsageValidateWorker(c *core.Context, interval time.Duration) cron.Worker {
	log := logger.With("module", "workers", "name", NameSessionUsageValidate)

	handlerFunc := func(ctx context.Context) error {
		// Retrieve session records from the database.
		query := map[string]any{
			"node_addr":    c.NodeAddr().String(),
			"service_type": c.Service().Type().String(),
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

				removePeer := false

				// Check if the session exceeds the maximum allowed bytes.
				maxBytes := item.GetMaxBytes()
				if !maxBytes.IsZero() && item.GetTotalBytes().GTE(maxBytes) {
					log.Debug("Marking peer for removing from service",
						"id", item.GetID(), "peer_id", item.GetPeerID(), "cause", "exceeds max bytes",
						"total_bytes", item.GetTotalBytes(), "max_bytes", item.GetMaxBytes(),
					)

					removePeer = true
				}

				// Check if the session exceeds the maximum allowed duration.
				maxDuration := item.GetMaxDuration()
				if maxDuration != 0 && item.GetDuration() >= maxDuration {
					log.Debug("Marking peer for removing from service",
						"id", item.GetID(), "peer_id", item.GetPeerID(), "cause", "exceeds max duration",
						"duration", item.GetDuration(), "max_duration", maxDuration,
					)

					removePeer = true
				}

				// If the session exceeded any limits, remove the associated peer.
				if removePeer {
					log.Debug("Removing peer from service", "id", item.GetID(), "peer_id", item.GetPeerID())

					if err := c.RemovePeerIfExists(jobCtx, item.GetPeerID()); err != nil {
						return fmt.Errorf("removing peer %q for session %d from service: %w", item.GetPeerID(), item.GetID(), err)
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
// This worker ensures sessions are active and consistent between the database and blockchain.
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

				removePeerFunc := func() error {
					// Ensure that only sessions of the current service type are validated.
					if item.GetServiceType() != c.Service().Type() {
						log.Debug("Skipping peer",
							"id", item.GetID(), "peer_id", item.GetPeerID(), "cause", "invalid service type",
							"got", item.GetServiceType(), "expected", c.Service().Type(),
						)

						return nil
					}

					remove := false

					// Remove peer if the session is missing on the blockchain.
					if session == nil {
						log.Debug("Marking peer for removing from service",
							"id", item.GetID(), "peer_id", item.GetPeerID(), "cause", "nil session",
						)

						remove = true
					}

					// Remove peer if the session status is not active.
					if session != nil && !session.GetStatus().Equal(v1.StatusActive) {
						log.Debug("Marking peer for removing from service",
							"id", item.GetID(), "peer_id", item.GetPeerID(), "cause", "invalid session status",
							"got", session.GetStatus(), "expected", v1.StatusActive,
						)

						remove = true
					}

					// Remove the associated peer if validation fails.
					if remove {
						log.Debug("Removing peer from service", "id", item.GetID(), "peer_id", item.GetPeerID())

						if err := c.RemovePeerIfExists(jobCtx, item.GetPeerID()); err != nil {
							return fmt.Errorf("removing peer %q for session %d from service: %w", item.GetPeerID(), item.GetID(), err)
						}
					}

					return nil
				}

				if err := removePeerFunc(); err != nil {
					return err
				}

				deleteSessionFunc := func() error {
					remove := false

					// Delete session if the session is missing on the blockchain.
					if session == nil {
						log.Debug("Marking session for deleting from database",
							"id", item.GetID(), "peer_id", item.GetPeerID(), "cause", "nil session",
						)

						remove = true
					}

					// Delete the session record from the database if not found on the blockchain.
					if remove {
						query := map[string]any{
							"id": item.GetID(),
						}

						log.Info("Deleting session from database", "id", item.GetID(), "peer_id", item.GetPeerID())

						if _, err := operations.SessionFindOneAndDelete(c.Database(), query); err != nil {
							return fmt.Errorf("deleting session %d from database: %w", item.GetID(), err)
						}
					}

					return nil
				}

				if err := deleteSessionFunc(); err != nil {
					return err
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
