package autoscaler

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/metrics"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

// shortenElection cannot go below a 1s lease: the Lease resource stores whole seconds, and a
// record rounded to zero looks expired to every candidate. Timings are restored when the test
// returns, so join every goroutine in the test body, not in t.Cleanup.
func shortenElection(t *testing.T) {
	t.Helper()

	saved := leaderTimings
	t.Cleanup(func() { leaderTimings = saved })

	leaderTimings.Lease = time.Second
	leaderTimings.Renew = 700 * time.Millisecond
	leaderTimings.Retry = 100 * time.Millisecond
}

type election struct {
	t      *testing.T
	client *fake.Clientset
	wg     sync.WaitGroup

	mu      sync.Mutex
	starts  map[string]int
	leader  string
	cancels map[string]context.CancelFunc
	errs    map[string]error
}

func newElection(t *testing.T) *election {
	t.Helper()
	shortenElection(t)

	return &election{
		t:       t,
		client:  fake.NewClientset(),
		starts:  make(map[string]int),
		cancels: make(map[string]context.CancelFunc),
		errs:    make(map[string]error),
	}
}

func (e *election) start(identity string) {
	ctx, cancel := context.WithCancel(context.Background())

	e.mu.Lock()
	e.cancels[identity] = cancel
	e.mu.Unlock()

	e.wg.Go(func() {
		err := RunLeader(ctx, e.client, testNamespace, identity, func(ctx context.Context) {
			e.mu.Lock()
			e.starts[identity]++
			e.leader = identity
			e.mu.Unlock()

			<-ctx.Done()

			e.mu.Lock()
			if e.leader == identity {
				e.leader = ""
			}
			e.mu.Unlock()
		})

		e.mu.Lock()
		e.errs[identity] = err
		e.mu.Unlock()
	})
}

func (e *election) stop(identity string) {
	e.mu.Lock()
	cancel := e.cancels[identity]
	e.mu.Unlock()

	require.NotNil(e.t, cancel, "no candidate %q", identity)
	cancel()
}

// shutdown must be deferred by every test so the goroutines are gone before the timings are restored.
func (e *election) shutdown() {
	e.mu.Lock()
	for _, cancel := range e.cancels {
		cancel()
	}
	e.mu.Unlock()

	e.wg.Wait()

	e.mu.Lock()
	defer e.mu.Unlock()

	for identity, err := range e.errs {
		require.NoError(e.t, err, "candidate %q", identity)
	}
}

func (e *election) startsFor(identity string) int {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.starts[identity]
}

func (e *election) currentLeader() string {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.leader
}

func TestRunLeaderRunsOnOneCandidateOnly(t *testing.T) {
	e := newElection(t)
	defer e.shutdown()

	e.start("a")
	e.start("b")

	// The gauge is read inside the poll: a term ending between the poll and a separate read would
	// fail on a value that was correct when it was written.
	require.Eventually(t, func() bool {
		return e.currentLeader() != "" && metricValue(t, metrics.AutoscalerIsLeader) == 1
	}, 5*leaderTimings.Lease, leaderTimings.Retry, "no candidate acquired the lease")

	leader := e.currentLeader()

	follower := "b"
	if leader == follower {
		follower = "a"
	}

	// Mutual exclusion is deliberately not asserted: client-go cannot guarantee it for an incumbent
	// descheduled past the renew deadline, which the 1s test lease makes likely under -race.
	// A stall past the lease duration hands the lease over for real, hence the guard.
	if e.currentLeader() != leader {
		t.Logf("the lease moved off %q before the assertion, mutual exclusion not checked", leader)

		return
	}

	require.Zero(t, e.startsFor(follower),
		"candidate %q ran while %q held the lease", follower, leader)
}

func TestRunLeaderHandsOverWhenLeaderStops(t *testing.T) {
	e := newElection(t)
	defer e.shutdown()

	e.start("a")
	e.start("b")

	require.Eventually(t, func() bool { return e.currentLeader() != "" },
		5*leaderTimings.Lease, leaderTimings.Retry, "no candidate acquired the lease")

	leader := e.currentLeader()
	loser := "b"

	if leader == loser {
		loser = "a"
	}

	e.stop(leader)

	require.Eventually(t, func() bool { return e.startsFor(loser) >= 1 },
		5*leaderTimings.Lease, leaderTimings.Retry, "candidate %q did not take over", loser)
	require.Equal(t, 1, e.startsFor(leader), "the stopped candidate ran again")
}

func TestRunLeaderRejectsEmptyIdentity(t *testing.T) {
	err := RunLeader(t.Context(), fake.NewClientset(), testNamespace, "", func(context.Context) {
		t.Error("run() called without an identity")
	})

	require.Error(t, err)
}

func TestRunLeaderReturnsOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan error, 1)

	go func() {
		done <- RunLeader(ctx, fake.NewClientset(), testNamespace, "a", func(context.Context) {
			t.Error("run() called after the context was cancelled")
		})
	}()

	select {
	case err := <-done:
		require.NoError(t, err)

	case <-time.After(time.Second):
		t.Fatal("RunLeader() did not return on a cancelled context")
	}
}

func TestRunLeaderReentersElectionAfterLosingLease(t *testing.T) {
	e := newElection(t)
	defer e.shutdown()

	var broken atomic.Bool

	e.client.PrependReactor("update", "leases", func(ktesting.Action) (bool, runtime.Object, error) {
		if broken.Load() {
			return true, nil, apierrors.NewInternalError(errors.New("lease update rejected"))
		}

		return false, nil, nil
	})

	e.start("a")

	require.Eventually(t, func() bool { return e.currentLeader() == "a" },
		5*leaderTimings.Lease, leaderTimings.Retry, "candidate did not acquire the lease")

	// Failing renewals end the term while the candidate's context stays alive.
	broken.Store(true)

	require.Eventually(t, func() bool { return e.currentLeader() == "" },
		5*leaderTimings.Lease, leaderTimings.Retry, "candidate kept the lease it could not renew")

	broken.Store(false)

	require.Eventually(t, func() bool { return e.startsFor("a") >= 2 },
		5*leaderTimings.Lease, leaderTimings.Retry, "candidate did not re-enter the election")
}

// client-go never joins the run callback, so only RunLeader keeps a tick that outlived its term
// from running next to the following term.
func TestRunLeaderWaitsForTheRunOfTheLostTerm(t *testing.T) {
	shortenElection(t)

	client := fake.NewClientset()

	var broken atomic.Bool

	client.PrependReactor("update", "leases", func(ktesting.Action) (bool, runtime.Object, error) {
		if broken.Load() {
			return true, nil, apierrors.NewInternalError(errors.New("lease update rejected"))
		}

		return false, nil, nil
	})

	var starts, running atomic.Int32

	var lost atomic.Bool

	release := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() {
		done <- RunLeader(ctx, client, testNamespace, "a", func(ctx context.Context) {
			starts.Add(1)

			if running.Add(1) > 1 {
				t.Error("run() started again while the run of the previous term was still going")
			}

			defer running.Add(-1)

			<-ctx.Done()
			lost.Store(true)

			// A tick outliving its term, like a probe waiting out the AMQP heartbeat.
			<-release
		})
	}()

	require.Eventually(t, func() bool { return starts.Load() >= 1 },
		5*leaderTimings.Lease, leaderTimings.Retry, "candidate did not acquire the lease")

	broken.Store(true)

	require.Eventually(t, lost.Load, 5*leaderTimings.Lease, leaderTimings.Retry,
		"candidate kept the lease it could not renew")

	broken.Store(false)

	require.Never(t, func() bool { return starts.Load() >= 2 },
		2*leaderTimings.Lease, leaderTimings.Retry, "a new term started before the previous run returned")

	close(release)

	require.Eventually(t, func() bool { return starts.Load() >= 2 },
		5*leaderTimings.Lease, leaderTimings.Retry, "candidate did not re-enter the election")

	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)

	case <-time.After(5 * leaderTimings.Lease):
		t.Fatal("RunLeader() did not return after its context was cancelled")
	}
}
