package autoscaler

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
)

// The cooldown outlasts a full down window many times over, so the candidate that scaled up cannot
// also scale down while the test body observes the action and takes its lease away. It still fits
// inside the take-over budget of 5 lease durations.
const (
	handoverInterval = 20 * time.Millisecond
	handoverCooldown = 2 * time.Second
	handoverPoll     = 5 * time.Millisecond
)

// leaderCandidate is one replica. Each has its own recorder, which is what tells the test who
// performed an update.
type leaderCandidate struct {
	identity string
	events   *record.FakeRecorder
	cancel   context.CancelFunc
	done     chan error
	stopped  bool

	mu       sync.Mutex
	recorded []string
}

func startCandidate(t *testing.T, client *fake.Clientset, identity string, cfg TargetConfig) *leaderCandidate {
	t.Helper()

	events := record.NewFakeRecorder(8)

	as := &Autoscaler{
		Targets:  []TargetConfig{cfg},
		Scaler:   &Scaler{Client: client, Recorder: events, Namespace: testNamespace},
		Interval: handoverInterval,
		Cooldown: handoverCooldown,
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := &leaderCandidate{
		identity: identity,
		events:   events,
		cancel:   cancel,
		done:     make(chan error, 1),
	}

	go func() {
		c.done <- RunLeader(ctx, client, testNamespace, identity, as.Run)
	}()

	return c
}

// stop waits for the candidate to leave the election, so later assertions cannot race with its
// ticks. Idempotent, so a test can stop one candidate early and still defer stopping both.
func (c *leaderCandidate) stop(t *testing.T) {
	t.Helper()

	if c.stopped {
		return
	}

	c.stopped = true

	c.cancel()

	select {
	case err := <-c.done:
		require.NoError(t, err, "candidate %q", c.identity)

	case <-time.After(5 * leaderTimings.Lease):
		t.Fatalf("candidate %q did not leave the election", c.identity)
	}
}

// collect returns every event recorded so far; safe to call from require.Eventually.
func (c *leaderCandidate) collect() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	for {
		select {
		case event := <-c.events.Events:
			c.recorded = append(c.recorded, event)

		default:
			return slices.Clone(c.recorded)
		}
	}
}

// settleOnUpdate rolls every updated Deployment out instantly; the fake client never touches
// status, so without it the health guard would block every action but the first.
func settleOnUpdate(client *fake.Clientset) {
	client.PrependReactor("update", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
		update, ok := action.(ktesting.UpdateAction)
		if !ok {
			return false, nil, nil
		}

		dep, ok := update.GetObject().(*appsv1.Deployment)
		if !ok {
			return false, nil, nil
		}

		dep.Status.AvailableReplicas = ptr.Deref(dep.Spec.Replicas, 1)
		dep.Status.ObservedGeneration = dep.Generation

		return false, nil, nil
	})
}

func TestAutoscalerHandsOverToTheNextLeader(t *testing.T) {
	shortenElection(t)

	client := fake.NewClientset(testDeployment(2))
	settleOnUpdate(client)

	// One up window of backlog, then idle: worth exactly one scale-up and one scale-down. Each
	// candidate replays its own script, so a lease bounce mid-window cannot leave the next leader
	// with nothing but the idle tail.
	steps := append(backlogSteps(3, 1500), proberStep{Backlog: 0, Consumers: 1})
	newTarget := func() TargetConfig { return stubTarget("leader-handover", newStubProber(steps...)) }

	a := startCandidate(t, client, "a", newTarget())
	b := startCandidate(t, client, "b", newTarget())

	defer a.stop(t)
	defer b.stop(t)

	// The event is recorded after the update, so its arrival means the action is complete.
	require.Eventually(t, func() bool { return len(a.collect())+len(b.collect()) > 0 },
		5*leaderTimings.Lease, handoverPoll, "no candidate scaled the deployment up")

	leader, follower := a, b
	if len(b.collect()) > 0 {
		leader, follower = b, a
	}

	require.Equal(t, int32(3), replicas(t, client))
	require.Equal(t,
		[]string{"Normal Scaled autoscaler: 2 -> 3 replicas, backlog 1500 above 1000 for 3 samples"},
		leader.collect())
	require.Empty(t, follower.collect(),
		"candidate %q scaled while %q held the lease", follower.identity, leader.identity)

	// stop joins RunLeader, so the leader's ticks are over before the follower is observed and the
	// scale-down below can only be the follower's. The cooldown alone would not guarantee it: the
	// annotation is truncated to whole seconds, so a short one can read as elapsed immediately.
	leader.stop(t)

	require.Eventually(t, func() bool { return len(follower.collect()) > 0 },
		5*leaderTimings.Lease, handoverPoll, "candidate %q did not take over", follower.identity)

	require.Equal(t, int32(2), replicas(t, client))
	require.Len(t, follower.collect(), 1)
	// The sample count depends on how long the hand-over took.
	require.Regexp(t, `^Normal Scaled autoscaler: 3 -> 2 replicas, backlog 0 below 10 for \d+ samples$`,
		follower.collect()[0])

	require.Len(t, leader.collect(), 1, "candidate %q scaled after losing the lease", leader.identity)
	require.Equal(t, 2, countUpdates(t, client), "the two candidates updated the deployment twice")
}
