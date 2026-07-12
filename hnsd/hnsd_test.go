package hnsd

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestSetupCreatesPrefixDir(t *testing.T) {
	// A nested, not-yet-existing path; hnsd itself refuses to start when the
	// prefix directory is missing, so Setup must create it.
	prefixDir := filepath.Join(t.TempDir(), "home", "hnsd")

	d := New("hnsd", "10.8.0.1:53", 8, 5, prefixDir)
	d.execFile = "sh" // stand-in binary so LookPath succeeds without hnsd installed

	if err := d.Setup(context.Background()); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	info, err := os.Stat(prefixDir)
	if err != nil {
		t.Fatalf("prefix directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("prefix path %q is not a directory", prefixDir)
	}
}

func TestCommandArgs(t *testing.T) {
	d := New("hnsd", "10.8.0.1:53", 8, 5, "/data/.dvpnx/hnsd")

	cmd := d.command(context.Background())

	want := []string{
		"hnsd",
		"--rs-host", "10.8.0.1:53",
		"--pool-size", "8",
		"--checkpoint",
		"--prefix", "/data/.dvpnx/hnsd",
	}
	if !slices.Equal(cmd.Args, want) {
		t.Fatalf("command args = %v, want %v", cmd.Args, want)
	}
}

func TestShouldRestart(t *testing.T) {
	tests := []struct {
		name        string
		maxRestarts int
		restarts    int
		want        bool
	}{
		{"disabled", 0, 0, false},
		{"unlimited", -1, 1 << 20, true},
		{"below limit", 5, 4, true},
		{"at limit", 5, 5, false},
		{"above limit", 5, 6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := New("hnsd", "10.8.0.1:53", 8, tt.maxRestarts, t.TempDir())

			if got := d.shouldRestart(tt.restarts); got != tt.want {
				t.Fatalf("shouldRestart(%d) with maxRestarts=%d = %v, want %v",
					tt.restarts, tt.maxRestarts, got, tt.want)
			}
		})
	}
}
