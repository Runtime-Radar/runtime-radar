package main

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/autoscaler"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/config"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
)

type stubProber struct{ queue string }

func (stubProber) InspectQueue(context.Context) (messages, consumers int, err error) {
	return 0, 0, nil
}

func stubProbers() (func(queue string) autoscaler.Prober, *[]string) {
	var queues []string

	return func(queue string) autoscaler.Prober {
		queues = append(queues, queue)

		return stubProber{queue: queue}
	}, &queues
}

func autoscalerConfig() *config.Config {
	return &config.Config{
		CSNamespace:        "runtime-radar",
		AutoscalerEnabled:  true,
		AutoscalerInterval: 20 * time.Second,
		AutoscalerCooldown: 2 * time.Minute,
		AutoscalerEventProcessor: &config.AutoscalerTarget{
			Name:             "event-processor",
			Enabled:          true,
			Deployment:       "event-processor",
			Queue:            "runtime_events",
			MinReplicas:      2,
			MaxReplicas:      10,
			ScaleUpBacklog:   1000,
			ScaleUpWindow:    4,
			ScaleDownBacklog: 10,
			ScaleDownWindow:  40,
		},
		AutoscalerHistoryAPI: &config.AutoscalerTarget{
			Name:             "history-api",
			Enabled:          true,
			Deployment:       "history-api",
			Queue:            "history_events",
			MinReplicas:      3,
			MaxReplicas:      7,
			ScaleUpBacklog:   20000,
			ScaleUpWindow:    5,
			ScaleDownBacklog: 11,
			ScaleDownWindow:  41,
		},
	}
}

func TestBuildAutoscalerAllTargets(t *testing.T) {
	cfg := autoscalerConfig()
	clientset := fake.NewClientset()
	newProber, queues := stubProbers()

	as := buildAutoscaler(cfg, clientset, newProber, nil)
	require.Equal(t, []string{"runtime_events", "history_events"}, *queues)

	require.Equal(t, []autoscaler.TargetConfig{
		{
			Name:        "event-processor",
			Deployment:  "event-processor",
			Queue:       "runtime_events",
			Min:         2,
			Max:         10,
			UpBacklog:   1000,
			UpWindow:    4,
			DownBacklog: 10,
			DownWindow:  40,
			Prober:      stubProber{queue: "runtime_events"},
		},
		{
			Name:        "history-api",
			Deployment:  "history-api",
			Queue:       "history_events",
			Min:         3,
			Max:         7,
			UpBacklog:   20000,
			UpWindow:    5,
			DownBacklog: 11,
			DownWindow:  41,
			Prober:      stubProber{queue: "history_events"},
		},
	}, as.Targets)

	require.Equal(t, 20*time.Second, as.Interval)
	require.Equal(t, 2*time.Minute, as.Cooldown)
	require.Equal(t, "runtime-radar", as.Scaler.Namespace)
	require.Same(t, clientset, as.Scaler.Client)
	require.Nil(t, as.Scaler.Recorder)
}

func TestBuildAutoscalerSkipsDisabledTarget(t *testing.T) {
	cfg := autoscalerConfig()
	cfg.AutoscalerHistoryAPI.Enabled = false

	newProber, queues := stubProbers()

	as := buildAutoscaler(cfg, fake.NewClientset(), newProber, nil)

	require.Len(t, as.Targets, 1)
	require.Equal(t, "event-processor", as.Targets[0].Name)
	require.Equal(t, stubProber{queue: "runtime_events"}, as.Targets[0].Prober)
	require.Equal(t, []string{"runtime_events"}, *queues)
}

// An unreachable broker must not keep the service from starting: the probers are wired anyway and
// the loop reports the failure sample by sample until the broker answers.
func TestRabbitProbersUnreachableBroker(t *testing.T) {
	cfg := autoscalerConfig()
	cfg.RabbitAddr = closedAddr(t)

	newProber, closeProbers := rabbitProbers(cfg)
	defer closeProbers()

	as := buildAutoscaler(cfg, fake.NewClientset(), newProber, nil)
	require.Len(t, as.Targets, 2)

	for _, target := range as.Targets {
		require.NotNil(t, target.Prober)

		_, _, err := target.Prober.InspectQueue(t.Context())
		require.Error(t, err)
	}
}

// The brokers are opened by the leader when it takes its first sample, so a replica that never
// leads leaves RabbitMQ alone.
func TestRabbitProbersDialOnFirstProbe(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = lis.Close() })

	dialed := make(chan struct{}, 1)

	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				return
			}

			select {
			case dialed <- struct{}{}:
			default:
			}

			_ = conn.Close()
		}
	}()

	cfg := autoscalerConfig()
	cfg.RabbitAddr = lis.Addr().String()

	newProber, closeProbers := rabbitProbers(cfg)
	defer closeProbers()

	as := buildAutoscaler(cfg, fake.NewClientset(), newProber, nil)
	require.Len(t, as.Targets, 2)
	require.Empty(t, dialed)

	// The listener is no broker, so the probe fails; that it was dialed at all is the point.
	_, _, err = as.Targets[0].Prober.InspectQueue(t.Context())
	require.Error(t, err)

	select {
	case <-dialed:
	case <-time.After(5 * time.Second):
		t.Fatal("the prober never dialed the broker")
	}
}

// closedAddr returns the address of a listener closed right away, so nothing can answer on it.
func closedAddr(t *testing.T) string {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	require.NoError(t, lis.Close())

	return lis.Addr().String()
}
