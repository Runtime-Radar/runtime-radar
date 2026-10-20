package monitor_test

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/cilium/tetragon/api/v1/tetragon"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/monitor"
	"google.golang.org/grpc"
)

const testVersion = "v1.7.0-test"

// sensorsStub answers GetVersion and nothing else, which is all NewTetra calls.
type sensorsStub struct {
	tetragon.UnimplementedFineGuidanceSensorsServer
}

func (sensorsStub) GetVersion(context.Context, *tetragon.GetVersionRequest) (*tetragon.GetVersionResponse, error) {
	return &tetragon.GetVersionResponse{Version: testVersion}, nil
}

// serveUnix starts a Tetragon gRPC stub listening on a Unix socket named name and returns its
// path. The directory is not t.TempDir(), whose path embeds the test name and can overshoot the
// ~100-byte limit on Unix socket paths.
func serveUnix(t *testing.T, name string) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "rm")
	if err != nil {
		t.Fatalf("can't create socket dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	path := filepath.Join(dir, name)

	lis, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("can't listen on unix socket %s: %v", path, err)
	}

	srv := grpc.NewServer()
	tetragon.RegisterFineGuidanceSensorsServer(srv, sensorsStub{})

	go func() {
		_ = srv.Serve(lis)
	}()

	t.Cleanup(srv.Stop)

	return path
}

// TestNewTetraUnixSocket checks the assumption the Helm charts rely on: the gRPC dialer used by
// NewTetra accepts a unix:///path/to/socket target natively, with no custom resolver or dialer.
func TestNewTetraUnixSocket(t *testing.T) {
	path := serveUnix(t, "t.sock")

	tetra, closeConn, err := monitor.NewTetra("unix://"+path, 1)
	if err != nil {
		t.Fatalf("NewTetra over unix socket: %v", err)
	}
	defer func() {
		if err := closeConn(); err != nil {
			t.Errorf("can't close tetragon connection: %v", err)
		}
	}()

	if tetra.Version != testVersion {
		t.Errorf("tetragon version: got %q, want %q", tetra.Version, testVersion)
	}
}
