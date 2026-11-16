package rabbit

import (
	"context"
	"errors"
	"testing"
	"time"
)

// The paths that need a live broker — a successful inspect, and the connection close that hands a
// failed inspect to the reconnect loop — are covered by
// kube-manager/tests/integration/rabbit_inspect_queue_test.go, gated on RABBIT_ADDR.

func TestInspectQueueNotInitialized(t *testing.T) {
	mb := &MessageBroker{}

	messages, consumers, err := mb.InspectQueue(t.Context())
	if !errors.Is(err, ErrNotInitialized) {
		t.Errorf("InspectQueue() error = %v, want %v", err, ErrNotInitialized)
	}

	if messages != 0 || consumers != 0 {
		t.Errorf("InspectQueue() = %d, %d, want 0, 0", messages, consumers)
	}
}

func TestInspectQueueCanceledContext(t *testing.T) {
	mb := &MessageBroker{}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	messages, consumers, err := mb.InspectQueue(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("InspectQueue() error = %v, want %v", err, context.Canceled)
	}

	if messages != 0 || consumers != 0 {
		t.Errorf("InspectQueue() = %d, %d, want 0, 0", messages, consumers)
	}
}

func TestInspectQueueReconnecting(t *testing.T) {
	mb := &MessageBroker{}

	// The reconnect goroutine holds the write lock for the whole outage.
	mb.Lock()
	defer mb.Unlock()

	done := make(chan error, 1)
	go func() {
		_, _, err := mb.InspectQueue(t.Context())
		done <- err
	}()

	select {
	case err := <-done:
		if !errors.Is(err, ErrReconnecting) {
			t.Errorf("InspectQueue() error = %v, want %v", err, ErrReconnecting)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("InspectQueue blocked while the write lock was held")
	}
}

// Close must not wait out an outage, or SIGTERM hangs until SIGKILL.
func TestCloseReconnecting(t *testing.T) {
	mb := &MessageBroker{}

	mb.Lock()
	defer mb.Unlock()

	done := make(chan error, 1)
	go func() {
		done <- mb.Close()
	}()

	select {
	case err := <-done:
		if !errors.Is(err, ErrReconnecting) {
			t.Errorf("Close() error = %v, want %v", err, ErrReconnecting)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Close blocked while the write lock was held")
	}

	if mb.closed.Load() {
		t.Error("Close() gave up on a reconnecting broker but still marked it closed")
	}
}

// A broker that never connected has nothing to close, but it must still stop the reconnect loop.
func TestCloseNotInitialized(t *testing.T) {
	mb := &MessageBroker{}

	if err := mb.Close(); !errors.Is(err, ErrNotInitialized) {
		t.Errorf("Close() error = %v, want %v", err, ErrNotInitialized)
	}

	if !mb.closed.Load() {
		t.Error("Close() left the broker open, so the reconnect loop would redial")
	}
}
