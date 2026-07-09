package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/sentinel-official/sentinel-go-sdk/v2/libs/cron"
	logger "github.com/sentinel-official/sentinel-go-sdk/v2/libs/log"
	"github.com/sentinel-official/sentinel-go-sdk/v2/libs/speedtest"

	"github.com/sentinel-official/sentinel-dvpnx/core"
)

const NameSpeedtest = "speedtest"

// NewSpeedtestWorker creates a worker that performs periodic speed tests and updates the context with the results.
// This worker measures the download and upload speeds and sets the results in the application's context.
func NewSpeedtestWorker(c *core.Context, interval time.Duration) cron.Worker {
	log := logger.With("module", "workers", "name", NameSpeedtest)

	// Handler function that performs the speed test and updates the context.
	handlerFunc := func(ctx context.Context) error {
		// Run the speed test to measure download and upload speeds.
		dlSpeed, ulSpeed, err := speedtest.Run(ctx)
		if err != nil {
			return fmt.Errorf("running speed test: %w", err)
		}

		log.Debug("Updating context", "dl_speed", dlSpeed, "ul_speed", ulSpeed)
		c.SetSpeedtestResults(dlSpeed, ulSpeed)

		return nil
	}

	// Initialize and return the worker.
	return cron.NewBasicWorker(NameSpeedtest).
		WithHandler(handlerFunc).
		WithInterval(interval)
}
