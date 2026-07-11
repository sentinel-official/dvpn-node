package workers

import (
	"context"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/sentinel-official/sentinel-go-sdk/v2/libs/cron"
	logger "github.com/sentinel-official/sentinel-go-sdk/v2/libs/log"

	"github.com/sentinel-official/sentinel-dvpnx/core"
)

const NameBestRPCAddr = "best_rpc_addr"

// NewBestRPCAddrWorker creates a worker that determines the best RPC address based on latency.
// This worker periodically measures the latency of available RPC addresses,
// sorts them in ascending order of latency, and updates the context.
func NewBestRPCAddrWorker(c *core.Context, interval time.Duration) cron.Worker {
	client := &http.Client{Timeout: 5 * time.Second}
	log := logger.With("module", "workers", "name", NameBestRPCAddr)

	// Handler function that measures RPC address latencies and updates the context.
	handlerFunc := func(ctx context.Context) error {
		addrs := c.RPCAddrs()                       // List of RPC addresses from the context.
		latencies := make(map[string]time.Duration) // Maps each address to its latency.
		mu := &sync.Mutex{}                         // Synchronizes access to shared resources.
		wg := &sync.WaitGroup{}                     // Ensures all goroutines complete.

		// Measure latency for each address concurrently.
		for _, addr := range addrs {
			wg.Add(1)

			go func(addr string) {
				defer wg.Done()

				endpoint, err := url.JoinPath(addr, "/status")
				if err != nil {
					return
				}

				// Create the HTTP GET request to the endpoint.
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
				if err != nil {
					return
				}

				// Record start time and perform HTTP GET request.
				start := time.Now()

				resp, err := client.Do(req)
				if err != nil {
					return
				}

				defer func() {
					_ = resp.Body.Close()
				}()

				// Skip if the response status is not HTTP 200 OK.
				if resp.StatusCode != http.StatusOK {
					return
				}

				// Calculate and record the latency.
				latency := time.Since(start)

				mu.Lock()
				defer mu.Unlock()

				latencies[addr] = latency
			}(addr)
		}

		// Wait for all goroutines to complete.
		wg.Wait()

		// Return early if no RPC addresses are available.
		if len(latencies) == 0 {
			return nil
		}

		// Sort the addresses by latency.
		addrs = make([]string, 0, len(latencies))
		for addr := range latencies {
			addrs = append(addrs, addr)
		}

		sort.Slice(addrs, func(i, j int) bool {
			return latencies[addrs[i]] < latencies[addrs[j]]
		})

		log.Debug("Updating context", "addrs", addrs)
		c.SetRPCAddrs(addrs)

		return nil
	}

	// Initialize and return the worker.
	return cron.NewBasicWorker(NameBestRPCAddr).
		WithHandler(handlerFunc).
		WithInterval(interval)
}
