package autoscaler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/metrics"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/utils/ptr"
)

// Every test uses its own target name so the global metric vectors keep one series per test.
func stubTarget(name string, prober Prober) TargetConfig {
	return TargetConfig{
		Name:        name,
		Deployment:  "event-processor",
		Queue:       "runtime_events",
		Min:         2,
		Max:         4,
		UpBacklog:   1000,
		UpWindow:    3,
		DownBacklog: 10,
		DownWindow:  4,
		Prober:      prober,
	}
}

// The states are seeded here because only Run allocates them, and most tests drive tick directly.
func testAutoscaler(client *fake.Clientset, clock func() time.Time, targets ...TargetConfig) *Autoscaler {
	return &Autoscaler{
		Targets:  targets,
		Scaler:   &Scaler{Client: client, Namespace: testNamespace},
		Interval: time.Millisecond,
		Cooldown: testCooldown,
		Clock:    clock,
		states:   make([]targetState, len(targets)),
	}
}

func replicas(t *testing.T, client *fake.Clientset) int32 {
	t.Helper()

	return replicasOf(t, client, "event-processor")
}

func replicasOf(t *testing.T, client *fake.Clientset, name string) int32 {
	t.Helper()

	dep, err := client.AppsV1().Deployments(testNamespace).Get(t.Context(), name, metav1.GetOptions{})
	require.NoError(t, err)

	return ptr.Deref(dep.Spec.Replicas, 1)
}

// settle marks the Deployment rolled out; the fake client never updates status itself.
func settle(t *testing.T, client *fake.Clientset) {
	t.Helper()

	deployments := client.AppsV1().Deployments(testNamespace)

	dep, err := deployments.Get(t.Context(), "event-processor", metav1.GetOptions{})
	require.NoError(t, err)

	dep.Status.AvailableReplicas = ptr.Deref(dep.Spec.Replicas, 1)
	dep.Status.ObservedGeneration = dep.Generation

	_, err = deployments.Update(t.Context(), dep, metav1.UpdateOptions{})
	require.NoError(t, err)
}

func metricValue(t *testing.T, metric prometheus.Metric) float64 {
	t.Helper()

	var out dto.Metric
	require.NoError(t, metric.Write(&out))

	if gauge := out.GetGauge(); gauge != nil {
		return gauge.GetValue()
	}

	return out.GetCounter().GetValue()
}

func scaleActions(cfg TargetConfig, direction string) prometheus.Counter {
	return metrics.AutoscalerScaleActionsCount.With(prometheus.Labels{
		metrics.Target: cfg.Name, metrics.Direction: direction,
	})
}

// metricDelta measures growth since the snapshot; the vectors are process globals.
func metricDelta(t *testing.T, metric prometheus.Metric) func() float64 {
	t.Helper()

	before := metricValue(t, metric)

	return func() float64 { return metricValue(t, metric) - before }
}

func TestTickScalesUpAfterWindow(t *testing.T) {
	client := fake.NewClientset(testDeployment(2))
	prober := newStubProber(backlogSteps(10, 1500)...)
	cfg := stubTarget("up-window", prober)
	as := testAutoscaler(client, func() time.Time { return testNow }, cfg)

	actions := metricDelta(t, scaleActions(cfg, directionUp))

	for i := range cfg.UpWindow - 1 {
		as.tick(t.Context())

		require.Equal(t, int32(2), replicas(t, client), "tick %d scaled before the window was full", i+1)
	}

	as.tick(t.Context())

	require.Equal(t, int32(3), replicas(t, client))
	require.Zero(t, as.states[0].high, "the high counter survived the action")
	require.Equal(t, float64(1), actions())
	require.Equal(t, float64(3), metricValue(t, replicasGauge(cfg)))
	require.Equal(t, float64(1500), metricValue(t, backlogGauge(cfg)))
}

func TestTickKeepsTargetsIndependent(t *testing.T) {
	client := fake.NewClientset(testDeployment(2), namedDeployment("history-api", 2))

	eventProcessor := stubTarget("multi-event-processor", newStubProber(backlogSteps(10, 1500)...))

	// A neutral backlog with a failed probe in the middle: no action, and gauges of its own.
	historySteps := append([]proberStep{{Backlog: 500, Consumers: 1}, {Err: errors.New("broker is reconnecting")}},
		backlogSteps(10, 500)...)

	historyAPI := stubTarget("multi-history-api", newStubProber(historySteps...))
	historyAPI.Deployment = "history-api"
	historyAPI.Queue = "history_events"

	as := testAutoscaler(client, func() time.Time { return testNow }, eventProcessor, historyAPI)

	upActions := metricDelta(t, scaleActions(eventProcessor, directionUp))
	historyActions := metricDelta(t, scaleActions(historyAPI, directionUp))

	for range eventProcessor.UpWindow {
		as.tick(t.Context())
	}

	require.Equal(t, int32(3), replicasOf(t, client, "event-processor"))
	require.Equal(t, int32(2), replicasOf(t, client, "history-api"), "the neutral target was scaled")
	require.Equal(t, float64(1), upActions())
	require.Zero(t, historyActions())

	require.Zero(t, as.states[0].high, "the action did not reset the counter of its own target")
	require.Zero(t, as.states[1].high)
	require.Zero(t, as.states[1].low)

	require.Equal(t, float64(1500), metricValue(t, backlogGauge(eventProcessor)))
	require.Equal(t, float64(500), metricValue(t, backlogGauge(historyAPI)),
		"the targets share a backlog gauge series")
	require.Equal(t, float64(3), metricValue(t, replicasGauge(eventProcessor)))
	require.Equal(t, float64(2), metricValue(t, replicasGauge(historyAPI)))
}

func TestTickReportsQueueConsumers(t *testing.T) {
	client := fake.NewClientset(testDeployment(2))
	cfg := stubTarget("consumers", newStubProber(
		proberStep{Backlog: 500, Consumers: 3},
		proberStep{Backlog: 500, Consumers: 0},
	))
	as := testAutoscaler(client, func() time.Time { return testNow }, cfg)

	as.tick(t.Context())

	require.Equal(t, float64(3), metricValue(t, consumersGauge(cfg)))

	// A queue nobody reads looks idle, so the count is the only sign of a wrong queue name.
	as.tick(t.Context())

	require.Equal(t, float64(0), metricValue(t, consumersGauge(cfg)))
	require.Equal(t, int32(2), replicasOf(t, client, "event-processor"), "the consumer count changed the decision")
}

func TestRunForgetsGaugesOnTermEnd(t *testing.T) {
	client := fake.NewClientset(testDeployment(2))
	cfg := stubTarget("forget-gauges", newStubProber(backlogSteps(10, 500)...))
	as := testAutoscaler(client, func() time.Time { return testNow }, cfg)

	as.tick(t.Context())

	require.Equal(t, float64(500), metricValue(t, backlogGauge(cfg)))

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	as.Run(ctx)

	// With returns a fresh series, so a forgotten gauge reads as zero rather than the last sample.
	require.Zero(t, metricValue(t, backlogGauge(cfg)), "an ex-leader kept exporting its last backlog")
	require.Zero(t, metricValue(t, consumersGauge(cfg)))
	require.Zero(t, metricValue(t, replicasGauge(cfg)))
}

func TestTickCooldownBlocksSecondScaleUp(t *testing.T) {
	client := fake.NewClientset(testDeployment(2))
	cfg := stubTarget("cooldown", newStubProber(backlogSteps(20, 1500)...))

	now := testNow
	as := testAutoscaler(client, func() time.Time { return now }, cfg)

	for range cfg.UpWindow {
		as.tick(t.Context())
	}

	require.Equal(t, int32(3), replicas(t, client), "the first window did not fire")

	settle(t, client)

	now = now.Add(testCooldown - time.Second)

	for range cfg.UpWindow {
		as.tick(t.Context())
	}

	require.Equal(t, int32(3), replicas(t, client), "the cooldown did not hold")

	now = now.Add(2 * time.Second)

	for range cfg.UpWindow {
		as.tick(t.Context())
	}

	require.Equal(t, int32(4), replicas(t, client), "the elapsed cooldown still held")
}

func TestTickScalesDownAfterWindow(t *testing.T) {
	client := fake.NewClientset(testDeployment(3))
	cfg := stubTarget("down-window", newStubProber(backlogSteps(10, 0)...))
	as := testAutoscaler(client, func() time.Time { return testNow }, cfg)

	actions := metricDelta(t, scaleActions(cfg, directionDown))

	for i := range cfg.DownWindow - 1 {
		as.tick(t.Context())

		require.Equal(t, int32(3), replicas(t, client), "tick %d scaled before the window was full", i+1)
	}

	as.tick(t.Context())

	require.Equal(t, int32(2), replicas(t, client))
	require.Zero(t, as.states[0].low, "the low counter survived the action")
	require.Equal(t, float64(1), actions())
}

func TestTickClampsDownFromAboveMax(t *testing.T) {
	client := fake.NewClientset(testDeployment(6))
	cfg := stubTarget("above-max", newStubProber(backlogSteps(2, 500)...))
	as := testAutoscaler(client, func() time.Time { return testNow }, cfg)

	actions := metricDelta(t, scaleActions(cfg, directionDown))

	as.tick(t.Context())

	require.Equal(t, cfg.Max, replicas(t, client))
	require.Equal(t, float64(1), actions())
}

func TestTickProberErrorResetsCounters(t *testing.T) {
	client := fake.NewClientset(testDeployment(2))

	steps := backlogSteps(2, 1500)
	steps = append(steps, proberStep{Err: errors.New("broker is reconnecting")})
	steps = append(steps, backlogSteps(3, 1500)...)

	cfg := stubTarget("prober-error", newStubProber(steps...))
	as := testAutoscaler(client, func() time.Time { return testNow }, cfg)

	// Without the reset, high-high-error-high would count as three consecutive samples.
	for range 3 {
		as.tick(t.Context())
	}

	require.Zero(t, as.states[0].high, "the failed probe did not break the window")
	require.Equal(t, float64(1500), metricValue(t, backlogGauge(cfg)),
		"the failed probe made the backlog gauge report an empty queue")

	as.tick(t.Context())

	require.Equal(t, int32(2), replicas(t, client), "the broken window still fired")

	for range cfg.UpWindow - 1 {
		as.tick(t.Context())
	}

	require.Equal(t, int32(3), replicas(t, client), "a full window of samples did not fire")
}

func TestTickUnhealthyDeploymentHoldsScaleUp(t *testing.T) {
	dep := testDeployment(2)
	dep.Status.AvailableReplicas = 1

	client := fake.NewClientset(dep)
	cfg := stubTarget("unhealthy-up", newStubProber(backlogSteps(10, 1500)...))
	as := testAutoscaler(client, func() time.Time { return testNow }, cfg)

	for range 2 * cfg.UpWindow {
		as.tick(t.Context())
	}

	require.Equal(t, int32(2), replicas(t, client), "an unhealthy deployment was scaled up")
	require.Zero(t, countUpdates(t, client))
}

func TestTickUnhealthyDeploymentScalesDown(t *testing.T) {
	dep := testDeployment(3)
	dep.Status.AvailableReplicas = 2

	client := fake.NewClientset(dep)
	cfg := stubTarget("unhealthy-down", newStubProber(backlogSteps(10, 0)...))
	as := testAutoscaler(client, func() time.Time { return testNow }, cfg)

	for range cfg.DownWindow {
		as.tick(t.Context())
	}

	require.Equal(t, int32(2), replicas(t, client), "an unhealthy deployment was not scaled down")
}

func TestTickGetErrorSkips(t *testing.T) {
	client := fake.NewClientset()
	prober := newStubProber(backlogSteps(10, 1500)...)
	as := testAutoscaler(client, func() time.Time { return testNow }, stubTarget("missing", prober))

	for range 4 {
		as.tick(t.Context())
	}

	require.Zero(t, countUpdates(t, client), "a missing deployment was updated")
	require.Equal(t, 4, prober.calls())
}

// Unlike a probe error, an unreadable deployment says nothing about the backlog, so the window survives.
func TestTickGetErrorKeepsCounters(t *testing.T) {
	client := fake.NewClientset(testDeployment(2))

	var gets int

	client.PrependReactor("get", "deployments", func(ktesting.Action) (bool, runtime.Object, error) {
		gets++
		if gets == 3 {
			return true, nil, apierrors.NewInternalError(errors.New("etcd is unavailable"))
		}

		return false, nil, nil
	})

	cfg := stubTarget("get-error", newStubProber(backlogSteps(10, 1500)...))
	as := testAutoscaler(client, func() time.Time { return testNow }, cfg)

	for range cfg.UpWindow + 1 {
		as.tick(t.Context())
	}

	require.Equal(t, int32(3), replicas(t, client), "the window did not survive the failed read")
	require.Equal(t, 1, countUpdates(t, client))
}

// A failed non-conflict update keeps the window full, so the next tick retries.
func TestTickApplyErrorKeepsCounters(t *testing.T) {
	client := fake.NewClientset(testDeployment(2))
	client.PrependReactor("update", "deployments", func(ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewInternalError(errors.New("etcd is unavailable"))
	})

	cfg := stubTarget("apply-error", newStubProber(backlogSteps(10, 1500)...))
	as := testAutoscaler(client, func() time.Time { return testNow }, cfg)

	actions := metricDelta(t, scaleActions(cfg, directionUp))

	for range cfg.UpWindow {
		as.tick(t.Context())
	}

	require.Equal(t, cfg.UpWindow, as.states[0].high, "the failed update reset the window")
	require.Equal(t, 1, countUpdates(t, client))

	as.tick(t.Context())

	require.Equal(t, 2, countUpdates(t, client), "the next tick did not retry the action")
	require.Equal(t, int32(2), replicas(t, client))
	require.Zero(t, actions(), "a failed update was counted as a scaling action")
}

func TestTickConflictDoesNotRetry(t *testing.T) {
	client := fake.NewClientset(testDeployment(2))
	client.PrependReactor("update", "deployments", func(ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewConflict(
			schema.GroupResource{Group: "apps", Resource: "deployments"},
			"event-processor",
			errors.New("the object has been modified"),
		)
	})

	cfg := stubTarget("conflict", newStubProber(backlogSteps(10, 1500)...))
	as := testAutoscaler(client, func() time.Time { return testNow }, cfg)

	for range cfg.UpWindow {
		as.tick(t.Context())
	}

	require.Equal(t, 1, countUpdates(t, client), "the update was retried within the tick")
	require.Equal(t, int32(2), replicas(t, client))
}

func TestTickNilReplicasClampedToMin(t *testing.T) {
	dep := testDeployment(1)
	dep.Spec.Replicas = nil

	client := fake.NewClientset(dep)
	cfg := stubTarget("nil-replicas", newStubProber(backlogSteps(2, 500)...))
	as := testAutoscaler(client, func() time.Time { return testNow }, cfg)

	actions := metricDelta(t, scaleActions(cfg, directionUp))

	as.tick(t.Context())

	require.Equal(t, cfg.Min, replicas(t, client))
	require.Equal(t, float64(1), actions())
}

func TestTickDefaultClock(t *testing.T) {
	client := fake.NewClientset(testDeployment(2))
	cfg := stubTarget("default-clock", newStubProber(backlogSteps(10, 1500)...))

	as := testAutoscaler(client, nil, cfg)

	for range cfg.UpWindow {
		as.tick(t.Context())
	}

	dep, err := client.AppsV1().Deployments(testNamespace).Get(t.Context(), "event-processor", metav1.GetOptions{})
	require.NoError(t, err)
	require.False(t, LastScaled(dep).IsZero(), "the action was not stamped with the wall clock time")
}

func TestRunResetsCountersBetweenTerms(t *testing.T) {
	client := fake.NewClientset(testDeployment(2))
	cfg := stubTarget("term-counters", newStubProber(backlogSteps(100, 1500)...))
	as := testAutoscaler(client, func() time.Time { return testNow }, cfg)

	// One sample short of the up window, as a term ending mid-window leaves it.
	for range cfg.UpWindow - 1 {
		as.tick(t.Context())
	}

	require.Equal(t, cfg.UpWindow-1, as.states[0].high)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	as.Run(ctx)

	require.Zero(t, as.states[0].high, "a new term started on the counters of the previous one")
}

func TestRunStopsOnContextCancel(t *testing.T) {
	client := fake.NewClientset(testDeployment(2))
	prober := newStubProber(backlogSteps(100, 500)...)
	as := testAutoscaler(client, func() time.Time { return testNow }, stubTarget("run", prober))

	ctx, cancel := context.WithCancel(t.Context())

	done := make(chan struct{})

	go func() {
		defer close(done)

		as.Run(ctx)
	}()

	require.Eventually(t, func() bool { return prober.calls() > 0 },
		time.Second, as.Interval, "the loop never ticked")

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run() did not return within a second of the context being cancelled")
	}
}
