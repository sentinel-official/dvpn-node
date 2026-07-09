package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/sentinel-official/sentinel-go-sdk/v2/libs/cron"
	"github.com/sentinel-official/sentinelhub/v12/types/v1"
	"github.com/sentinel-official/sentinelhub/v12/x/node/types/v3"

	"github.com/sentinel-official/sentinel-dvpnx/core"
)

const (
	NameNodeStatusUpdate = "node_status_update"
	NameNodePricesUpdate = "node_prices_update"
)

// NewNodeStatusUpdateWorker creates a worker to periodically update the node's status to active on the blockchain.
// This worker broadcasts a transaction to mark the node as active at regular intervals.
func NewNodeStatusUpdateWorker(c *core.Context, interval time.Duration) cron.Worker {
	// Handler function that updates the node's status to active.
	handlerFunc := func(ctx context.Context) error {
		// Create a message to update the node's status to active.
		msg := v3.NewMsgUpdateNodeStatusRequest(
			c.AccAddr().Bytes(),
			v1.StatusActive,
		)

		// Broadcast the transaction message to the blockchain.
		if err := c.BroadcastTx(ctx, msg); err != nil {
			return fmt.Errorf("broadcasting tx with update_node_status msg: %w", err)
		}

		return nil
	}

	// Initialize and return the worker.
	return cron.NewBasicWorker(NameNodeStatusUpdate).
		WithHandler(handlerFunc).
		WithInterval(interval).
		WithRetryDelay(5 * time.Second)
}

// NewNodePricesUpdateWorker creates a worker that periodically updates the node's prices on the blockchain.
// The worker computes the current quote prices using the OracleClient and broadcasts a MsgUpdateNodeDetailsRequest.
func NewNodePricesUpdateWorker(c *core.Context, interval time.Duration) cron.Worker {
	handlerFunc := func(ctx context.Context) (err error) {
		client := c.OracleClient()
		if client == nil {
			return nil
		}

		var (
			prices         v1.Prices
			gigabytePrices v1.Prices
			hourlyPrices   v1.Prices
		)

		prices, err = c.SanitizedGigabytePrices(ctx)
		if err != nil {
			return fmt.Errorf("sanitizing gigabyte prices: %w", err)
		}

		for _, price := range prices {
			price, err := price.UpdateQuoteValue(ctx, client.GetQuotePrice)
			if err != nil {
				return fmt.Errorf("updating quote price for denom %q: %w", price.Denom, err)
			}

			gigabytePrices = gigabytePrices.Add(price)
		}

		prices, err = c.SanitizedHourlyPrices(ctx)
		if err != nil {
			return fmt.Errorf("sanitizing hourly prices: %w", err)
		}

		for _, price := range prices {
			price, err := price.UpdateQuoteValue(ctx, client.GetQuotePrice)
			if err != nil {
				return fmt.Errorf("updating quote price for denom %q: %w", price.Denom, err)
			}

			hourlyPrices = hourlyPrices.Add(price)
		}

		// Construct the message to update node details with new prices.
		msg := v3.NewMsgUpdateNodeDetailsRequest(
			c.AccAddr().Bytes(),
			gigabytePrices,
			hourlyPrices,
			nil,
		)

		// Broadcast the transaction message to the blockchain.
		if err := c.BroadcastTx(ctx, msg); err != nil {
			return fmt.Errorf("broadcasting tx with update_node_details msg: %w", err)
		}

		return nil
	}

	// Initialize and return the worker.
	return cron.NewBasicWorker(NameNodePricesUpdate).
		WithHandler(handlerFunc).
		WithInterval(interval).
		WithRetryDelay(5 * time.Second)
}
