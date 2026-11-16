package autoscaler

import (
	"context"
	"sync"
)

var _ Prober = (*stubProber)(nil)

type proberStep struct {
	Backlog   int
	Consumers int
	Err       error
}

// stubProber replays steps and keeps returning the last one once the script runs out.
type stubProber struct {
	mu    sync.Mutex
	steps []proberStep
	count int
}

func newStubProber(steps ...proberStep) *stubProber {
	return &stubProber{steps: steps}
}

// backlogSteps report one consumer, so only the tests that ask for it hit the no-consumer warning.
func backlogSteps(count, backlog int) []proberStep {
	steps := make([]proberStep, count)
	for i := range steps {
		steps[i] = proberStep{Backlog: backlog, Consumers: 1}
	}

	return steps
}

func (p *stubProber) InspectQueue(context.Context) (messages, consumers int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.steps) == 0 {
		return 0, 0, nil
	}

	p.count++
	step := p.steps[min(p.count-1, len(p.steps)-1)]

	if step.Err != nil {
		return 0, 0, step.Err
	}

	return step.Backlog, step.Consumers, nil
}

func (p *stubProber) calls() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.count
}
