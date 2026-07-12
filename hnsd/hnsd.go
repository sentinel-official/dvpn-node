package hnsd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/sentinel-official/sentinel-go-sdk/v2/libs/log"
	"github.com/sentinel-official/sentinel-go-sdk/v2/process"
)

// restartBackoff is the fixed delay applied before restarting a crashed hnsd process.
const restartBackoff = 1 * time.Second

// Daemon supervises the hnsd Handshake DNS resolver as an external process.
type Daemon struct {
	*process.Manager // Embedded process manager for handling lifecycle.

	execFile    string // execFile is the hnsd binary name resolved from PATH.
	rsHost      string // rsHost is the recursive resolver bind address (host:port).
	poolSize    uint   // poolSize is the number of Handshake DNS peers.
	maxRestarts int    // maxRestarts bounds restarts (0 disables, -1 unlimited).
	prefixDir   string // prefixDir is the directory where hnsd persists chain state.
}

// New creates a new hnsd Daemon bound to the given recursive resolver host.
func New(name, rsHost string, poolSize uint, maxRestarts int, prefixDir string) *Daemon {
	return &Daemon{
		Manager:     process.NewManager(name),
		execFile:    "hnsd",
		rsHost:      rsHost,
		poolSize:    poolSize,
		maxRestarts: maxRestarts,
		prefixDir:   prefixDir,
	}
}

// Setup verifies the hnsd binary is available and creates the prefix directory
// before the Daemon is started. hnsd refuses to start if the prefix directory
// does not exist.
func (d *Daemon) Setup(ctx context.Context) error {
	return d.Manager.Setup(ctx, func() error { //nolint:wrapcheck
		if _, err := exec.LookPath(d.execFile); err != nil {
			return fmt.Errorf("looking up %q binary: %w", d.execFile, err)
		}

		if err := os.MkdirAll(d.prefixDir, 0o700); err != nil {
			return fmt.Errorf("creating prefix directory %q: %w", d.prefixDir, err)
		}

		return nil
	})
}

// Start launches the hnsd supervisor and returns the managed context.
func (d *Daemon) Start(parent context.Context) (context.Context, error) {
	return d.Manager.Start(parent, func(ctx context.Context) error { //nolint:wrapcheck
		d.Go(ctx, func() error {
			return d.supervise(ctx)
		})

		return nil
	})
}

// Wait blocks until the hnsd supervisor exits or the context is cancelled.
func (d *Daemon) Wait(ctx context.Context) error {
	return d.Manager.Wait(ctx, nil) //nolint:wrapcheck
}

// Stop cancels the supervisor context, gracefully terminating hnsd.
func (d *Daemon) Stop() error {
	return d.Manager.Stop(nil) //nolint:wrapcheck
}

// supervise runs hnsd and restarts it on unexpected exit, bounded by maxRestarts.
func (d *Daemon) supervise(ctx context.Context) error {
	restarts := 0

	for {
		log.Info("Starting hnsd", "rs_host", d.rsHost, "pool_size", d.poolSize)

		err := d.command(ctx).Run()

		// A cancelled context is a normal shutdown, not a crash.
		if ctx.Err() != nil {
			return nil //nolint:nilerr
		}

		if !d.shouldRestart(restarts) {
			return fmt.Errorf("hnsd exited after %d restart(s): %w", restarts, err)
		}

		restarts++

		log.Error("hnsd exited, restarting",
			"restart", restarts, "max_restarts", d.maxRestarts, "error", err,
		)

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(restartBackoff):
		}
	}
}

// shouldRestart reports whether a crashed process should be restarted given the
// number of restarts already performed.
func (d *Daemon) shouldRestart(restarts int) bool {
	if d.maxRestarts < 0 {
		return true
	}

	return restarts < d.maxRestarts
}

// command builds the hnsd process bound to the configured resolver host.
// Only --rs-host is pinned; the authoritative root nameserver keeps its loopback default.
// --checkpoint starts the initial sync from the hard-coded checkpoint instead of
// genesis, and --prefix persists chain state across restarts so subsequent syncs
// resume from the stored tip.
func (d *Daemon) command(ctx context.Context) *exec.Cmd {
	cmd := exec.CommandContext(
		ctx,
		d.execFile,
		"--rs-host", d.rsHost,
		"--pool-size", strconv.FormatUint(uint64(d.poolSize), 10),
		"--checkpoint",
		"--prefix", d.prefixDir,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = 5 * time.Second

	return cmd
}
