package autoscaler

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// Prober reports the state of a target's queue: the messages ready for delivery that drive the
// scaling, and the consumers attached to it, which tell an idle queue from a queue nobody reads.
type Prober interface {
	InspectQueue(ctx context.Context) (messages, consumers int, err error)
}

// ProberCloser is a Prober holding a connection of its own, such as a message broker.
type ProberCloser interface {
	Prober
	io.Closer
}

// LazyProber opens its underlying prober on the first probe and redials on every later one until
// that succeeds. Dialing the broker while the service starts would tie the pod and node inventory
// and KillPod — neither of which needs a broker — to its availability.
type LazyProber struct {
	Dial func() (ProberCloser, error)

	mu     sync.Mutex
	prober ProberCloser
}

var _ ProberCloser = (*LazyProber)(nil)

// InspectQueue dials on demand; the caller treats a dial failure like any other probe failure and
// tries again on the next tick.
func (p *LazyProber) InspectQueue(ctx context.Context) (messages, consumers int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.prober == nil {
		prober, err := p.Dial()
		if err != nil {
			return 0, 0, fmt.Errorf("open queue prober: %w", err)
		}

		p.prober = prober
	}

	return p.prober.InspectQueue(ctx)
}

// Close releases the underlying prober, and is a no-op when it was never dialed.
func (p *LazyProber) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.prober == nil {
		return nil
	}

	prober := p.prober
	p.prober = nil

	return prober.Close()
}
