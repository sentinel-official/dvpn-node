package node

import (
	"context"
	"fmt"

	"github.com/sentinel-official/sentinel-go-sdk/libs/cmux"
	"github.com/sentinel-official/sentinel-go-sdk/libs/cron"
	"github.com/sentinel-official/sentinel-go-sdk/libs/log"
	"github.com/sentinel-official/sentinel-go-sdk/process"
	"github.com/sentinel-official/sentinelhub/v12/x/node/types/v3"
	"golang.org/x/sync/errgroup"

	"github.com/sentinel-official/sentinel-dvpnx/core"
	"github.com/sentinel-official/sentinel-dvpnx/hnsd"
)

// Node represents the application node, holding its context, scheduler, and server.
type Node struct {
	*process.Manager // Embedded process manager for handling lifecycle.

	ctx          *core.Context   // Application code context.
	handshakeDNS *hnsd.Daemon    // Daemon supervising the hnsd Handshake DNS resolver.
	scheduler    *cron.Scheduler // Scheduler for managing periodic tasks.
	server       *cmux.Server    // HTTP server for handling API requests.
}

// New creates a new Node with the provided context.
func New(name string) *Node {
	return &Node{
		Manager: process.NewManager(name),
	}
}

// WithContext sets the core context for the Node and returns the updated Node.
func (n *Node) WithContext(ctx *core.Context) *Node {
	n.ctx = ctx

	return n
}

// HandshakeDNS returns the hnsd daemon configured for the Node, or nil when disabled.
func (n *Node) HandshakeDNS() *hnsd.Daemon {
	return n.handshakeDNS
}

// WithHandshakeDNS sets the hnsd daemon for the Node and returns the updated Node.
func (n *Node) WithHandshakeDNS(v *hnsd.Daemon) *Node {
	n.handshakeDNS = v

	return n
}

// WithScheduler sets the scheduler for the Node and returns the updated Node.
func (n *Node) WithScheduler(v *cron.Scheduler) *Node {
	n.scheduler = v

	return n
}

// WithServer sets the server for the Node and returns the updated Node.
func (n *Node) WithServer(v *cmux.Server) *Node {
	n.server = v

	return n
}

// Context returns the core context configured for the Node.
func (n *Node) Context() *core.Context {
	return n.ctx
}

// Scheduler returns the scheduler configured for the Node.
func (n *Node) Scheduler() *cron.Scheduler {
	return n.scheduler
}

// Server returns the server configured for the Node.
func (n *Node) Server() *cmux.Server {
	return n.server
}

// Register registers the node on the network if not already registered.
func (n *Node) Register(ctx context.Context) error {
	// Query the network to check if the node is already registered.
	node, err := n.Context().Client().Node(ctx, n.Context().NodeAddr())
	if err != nil {
		return fmt.Errorf("failed to query node: %w", err)
	}

	if node != nil {
		log.Info("Node already registered", "addr", n.Context().NodeAddr())

		return nil
	}

	gigabytePrices, err := n.Context().SanitizedGigabytePrices(ctx)
	if err != nil {
		return fmt.Errorf("sanitizing gigabyte prices: %w", err)
	}

	hourlyPrices, err := n.Context().SanitizedHourlyPrices(ctx)
	if err != nil {
		return fmt.Errorf("sanitizing hourly prices: %w", err)
	}

	log.Info("Registering node",
		"gigabyte_prices", gigabytePrices,
		"hourly_price", hourlyPrices,
		"remote_addrs", n.Context().APIAddrs(),
	)

	// Prepare a message to register the node.
	msg := v3.NewMsgRegisterNodeRequest(
		n.Context().AccAddr(),
		gigabytePrices,
		hourlyPrices,
		n.Context().APIAddrs(),
	)

	// Broadcast the registration transaction.
	if err := n.Context().BroadcastTx(ctx, msg); err != nil {
		return fmt.Errorf("broadcasting tx with register_node msg: %w", err)
	}

	log.Info("Node registered successfully", "addr", n.Context().NodeAddr())

	return nil
}

// UpdateDetails updates the node's pricing and address details on the network.
func (n *Node) UpdateDetails(ctx context.Context) error {
	gigabytePrices, err := n.Context().SanitizedGigabytePrices(ctx)
	if err != nil {
		return fmt.Errorf("sanitizing gigabyte prices: %w", err)
	}

	hourlyPrices, err := n.Context().SanitizedHourlyPrices(ctx)
	if err != nil {
		return fmt.Errorf("sanitizing hourly prices: %w", err)
	}

	log.Info("Updating node details",
		"gigabyte_prices", gigabytePrices,
		"hourly_prices", hourlyPrices,
		"remote_addrs", n.Context().APIAddrs(),
	)

	// Prepare a message to update the node's details.
	msg := v3.NewMsgUpdateNodeDetailsRequest(
		n.Context().NodeAddr(),
		gigabytePrices,
		hourlyPrices,
		n.Context().APIAddrs(),
	)

	// Broadcast the update transaction.
	if err := n.Context().BroadcastTx(ctx, msg); err != nil {
		return fmt.Errorf("broadcasting tx with update_node_details msg: %w", err)
	}

	log.Info("Node details updated successfully", "addr", n.Context().NodeAddr())

	return nil
}

// Start initializes the Node's services, scheduler, and API server.
func (n *Node) Start(ctx context.Context) (context.Context, error) {
	return n.Manager.Start(ctx, func(ctx context.Context) error { //nolint:contextcheck,wrapcheck
		if err := n.Register(ctx); err != nil {
			return fmt.Errorf("registering node: %w", err)
		}

		if err := n.UpdateDetails(ctx); err != nil {
			return fmt.Errorf("updating details: %w", err)
		}

		var (
			schedulerCtx context.Context
			serverCtx    context.Context
			serviceCtx   context.Context
		)

		sg := &errgroup.Group{}

		sg.Go(func() (err error) {
			log.Info("Starting scheduler")

			if schedulerCtx, err = n.Scheduler().Start(ctx); err != nil {
				return fmt.Errorf("starting scheduler: %w", err)
			}

			return nil
		})

		sg.Go(func() (err error) {
			log.Info("Starting API server")

			if serverCtx, err = n.Server().Start(ctx); err != nil {
				return fmt.Errorf("starting API server: %w", err)
			}

			return nil
		})

		sg.Go(func() (err error) {
			log.Info("Starting service")

			if serviceCtx, err = n.Context().Service().Start(ctx); err != nil {
				return fmt.Errorf("starting service: %w", err)
			}

			return nil
		})

		if err := sg.Wait(); err != nil {
			return fmt.Errorf("starting group: %w", err)
		}

		if n.HandshakeDNS() != nil {
			log.Info("Starting Handshake DNS")

			hnsdCtx, err := n.HandshakeDNS().Start(ctx)
			if err != nil {
				return fmt.Errorf("starting Handshake DNS: %w", err)
			}

			n.Go(ctx, func() error {
				if err := n.HandshakeDNS().Wait(hnsdCtx); err != nil {
					return fmt.Errorf("waiting Handshake DNS: %w", err)
				}

				return nil
			})
		}

		n.Go(ctx, func() error {
			if err := n.Scheduler().Wait(schedulerCtx); err != nil {
				return fmt.Errorf("waiting scheduler: %w", err)
			}

			return nil
		})

		n.Go(ctx, func() error {
			if err := n.Server().Wait(serverCtx); err != nil {
				return fmt.Errorf("waiting API server: %w", err)
			}

			return nil
		})

		n.Go(ctx, func() error {
			if err := n.Context().Service().Wait(serviceCtx); err != nil {
				return fmt.Errorf("waiting service: %w", err)
			}

			return nil
		})

		return nil
	})
}

// Wait blocks until all background goroutines launched exit.
func (n *Node) Wait(ctx context.Context) error {
	return n.Manager.Wait(ctx, nil) //nolint:wrapcheck
}

// Stop gracefully stops the Node's operations.
func (n *Node) Stop() error {
	return n.Manager.Stop(func() error { //nolint:wrapcheck
		sg := &errgroup.Group{}

		sg.Go(func() error {
			log.Info("Stopping scheduler")

			if err := n.Scheduler().Stop(); err != nil {
				return fmt.Errorf("stopping scheduler: %w", err)
			}

			return nil
		})

		sg.Go(func() error {
			log.Info("Stopping API server")

			if err := n.Server().Stop(); err != nil {
				return fmt.Errorf("stopping API server: %w", err)
			}

			return nil
		})

		sg.Go(func() error {
			log.Info("Stopping service")

			if err := n.Context().Service().Stop(); err != nil {
				return fmt.Errorf("stopping service: %w", err)
			}

			return nil
		})

		if n.HandshakeDNS() != nil {
			sg.Go(func() error {
				log.Info("Stopping Handshake DNS")

				if err := n.HandshakeDNS().Stop(); err != nil {
					return fmt.Errorf("stopping Handshake DNS: %w", err)
				}

				return nil
			})
		}

		if err := sg.Wait(); err != nil {
			return fmt.Errorf("stopping group: %w", err)
		}

		return nil
	})
}

// Cleanup cleans up resources used by the node.
func (n *Node) Cleanup() error {
	return n.Manager.Cleanup(nil) //nolint:wrapcheck
}
