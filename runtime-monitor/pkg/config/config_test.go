package config_test

import (
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/config"
)

// cfg is built exactly once per test binary: config.New registers its flags on the global
// flag.CommandLine and calls flag.Parse, so calling it a second time panics with "flag redefined".
var cfg *config.Config

func TestMain(m *testing.M) {
	// TETRAGON_ADDR, when set, becomes the flag default, so clear it to observe the
	// compiled-in one.
	_ = os.Unsetenv("TETRAGON_ADDR")

	cfg = config.New()

	os.Exit(m.Run())
}

// TestTetragonAddrDefault pins the CVE-2026-65960 contract: the Tetragon target the service falls
// back to must be a Unix socket, both as the exported constant and as the registered default of
// the tetragonAddr flag — the latter is what a deployment without TETRAGON_ADDR actually dials.
func TestTetragonAddrDefault(t *testing.T) {
	const wantPrefix = "unix://"

	if !strings.HasPrefix(config.DefaultTetragonAddr, wantPrefix) {
		t.Errorf("DefaultTetragonAddr = %q, want a %s target: TCP is not supported in Kubernetes (CVE-2026-65960)",
			config.DefaultTetragonAddr, wantPrefix)
	}

	f := flag.Lookup("tetragonAddr")
	if f == nil {
		t.Fatal("config.New did not register the tetragonAddr flag")
	}

	if f.DefValue != config.DefaultTetragonAddr {
		t.Errorf("tetragonAddr flag default = %q, want %q", f.DefValue, config.DefaultTetragonAddr)
	}

	if cfg.TetragonAddr != config.DefaultTetragonAddr {
		t.Errorf("Config.TetragonAddr = %q, want %q", cfg.TetragonAddr, config.DefaultTetragonAddr)
	}
}
