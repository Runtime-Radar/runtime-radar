package metrics

import (
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestPrepareRegistryGathersAutoscalerMetrics(t *testing.T) {
	t.Cleanup(func() {
		AutoscalerQueueBacklog.Reset()
		AutoscalerQueueConsumers.Reset()
		AutoscalerReplicas.Reset()
		AutoscalerScaleActionsCount.Reset()
		AutoscalerIsLeader.Set(0)
	})

	// Vectors only report the label combinations they have seen.
	AutoscalerQueueBacklog.WithLabelValues("event-processor", "runtime_events").Set(42)
	AutoscalerQueueConsumers.WithLabelValues("event-processor", "runtime_events").Set(2)
	AutoscalerReplicas.WithLabelValues("event-processor").Set(3)
	AutoscalerScaleActionsCount.WithLabelValues("event-processor", "up").Inc()
	AutoscalerIsLeader.Set(1)

	registry, err := PrepareRegistry("kube-manager", "test")
	require.NoError(t, err)

	families, err := registry.Gather()
	require.NoError(t, err)

	gathered := make(map[string]*dto.MetricFamily, len(families))
	for _, family := range families {
		gathered[family.GetName()] = family
	}

	tests := []struct {
		name   string
		labels map[string]string
		value  float64
	}{
		{
			name:   "autoscaler_queue_backlog",
			labels: map[string]string{"target": "event-processor", "queue": "runtime_events"},
			value:  42,
		},
		{
			name:   "autoscaler_queue_consumers",
			labels: map[string]string{"target": "event-processor", "queue": "runtime_events"},
			value:  2,
		},
		{
			name:   "autoscaler_replicas",
			labels: map[string]string{"target": "event-processor"},
			value:  3,
		},
		{
			name:   "autoscaler_scale_actions_count",
			labels: map[string]string{"target": "event-processor", "direction": "up"},
			value:  1,
		},
		{
			name:   "autoscaler_is_leader",
			labels: map[string]string{},
			value:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			family := gathered[tt.name]
			require.NotNil(t, family, "Gather() is missing metric %q", tt.name)
			require.Len(t, family.GetMetric(), 1)

			metric := family.GetMetric()[0]

			// The registry adds the mandatory service and cluster labels to every series.
			labels := map[string]string{}
			for _, pair := range metric.GetLabel() {
				labels[pair.GetName()] = pair.GetValue()
			}

			for name, want := range tt.labels {
				require.Equal(t, want, labels[name], "label %q of %q", name, tt.name)
			}

			require.Equal(t, "kube-manager", labels["service"])
			require.Equal(t, "test", labels["cluster"])
			require.Len(t, labels, len(tt.labels)+2, "unexpected labels on %q: %v", tt.name, labels)

			if counter := metric.GetCounter(); counter != nil {
				require.Equal(t, tt.value, counter.GetValue())

				return
			}

			require.NotNil(t, metric.GetGauge(), "%q is neither a gauge nor a counter", tt.name)
			require.Equal(t, tt.value, metric.GetGauge().GetValue())
		})
	}
}
