package config

import (
	"flag"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// newTargets uses a throwaway flag set; re-registering on flag.CommandLine panics.
func newTargets(t *testing.T) (eventProcessor, historyAPI *AutoscalerTarget) {
	t.Helper()

	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	return registerAutoscalerTarget(fs, eventProcessorEnvPrefix, eventProcessorFlagPrefix, eventProcessorDefaults()),
		registerAutoscalerTarget(fs, historyAPIEnvPrefix, historyAPIFlagPrefix, historyAPIDefaults())
}

func newConfig(t *testing.T) *Config {
	t.Helper()

	eventProcessor, historyAPI := newTargets(t)

	return &Config{
		CSNamespace:              "runtime-radar",
		AutoscalerEnabled:        true,
		AutoscalerInterval:       15 * time.Second,
		AutoscalerCooldown:       90 * time.Second,
		AutoscalerEventProcessor: eventProcessor,
		AutoscalerHistoryAPI:     historyAPI,
	}
}

func TestValidateAutoscalerDefaults(t *testing.T) {
	require.NoError(t, newConfig(t).ValidateAutoscaler())
}

func TestValidateAutoscalerDisabled(t *testing.T) {
	c := newConfig(t)
	c.AutoscalerEnabled = false
	c.CSNamespace = ""
	c.AutoscalerEventProcessor.Deployment = ""
	c.AutoscalerEventProcessor.Queue = ""
	c.AutoscalerEventProcessor.MinReplicas = -5
	c.AutoscalerEventProcessor.MaxReplicas = -10
	c.AutoscalerEventProcessor.ScaleUpWindow = 0
	c.AutoscalerEventProcessor.ScaleDownWindow = 0
	c.AutoscalerEventProcessor.ScaleUpBacklog = -1
	c.AutoscalerEventProcessor.ScaleDownBacklog = -1

	require.NoError(t, c.ValidateAutoscaler())
}

func TestValidateAutoscalerDisabledTargetNotValidated(t *testing.T) {
	c := newConfig(t)
	c.AutoscalerHistoryAPI.Enabled = false
	c.AutoscalerHistoryAPI.Queue = ""
	c.AutoscalerHistoryAPI.MinReplicas = 0

	require.NoError(t, c.ValidateAutoscaler())
}

func TestValidateAutoscalerAllTargetsDisabled(t *testing.T) {
	c := newConfig(t)
	c.AutoscalerEventProcessor.Enabled = false
	c.AutoscalerHistoryAPI.Enabled = false

	err := c.ValidateAutoscaler()
	require.ErrorContains(t, err, "at least one autoscaler target must be enabled")
}

func TestValidateAutoscalerNoNamespace(t *testing.T) {
	c := newConfig(t)
	c.CSNamespace = ""

	err := c.ValidateAutoscaler()
	require.ErrorContains(t, err, "CS_NAMESPACE")
}

func TestValidateAutoscalerInterval(t *testing.T) {
	c := newConfig(t)
	c.AutoscalerInterval = 0

	err := c.ValidateAutoscaler()
	require.ErrorContains(t, err, "AUTOSCALER_INTERVAL must be > 0")
}

func TestValidateAutoscalerCooldown(t *testing.T) {
	c := newConfig(t)
	c.AutoscalerCooldown = -time.Second

	err := c.ValidateAutoscaler()
	require.ErrorContains(t, err, "AUTOSCALER_COOLDOWN must be >= 0")
}

func TestValidateAutoscalerRules(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(target *AutoscalerTarget)
		want   string
	}{
		{
			name:   "empty deployment",
			mutate: func(target *AutoscalerTarget) { target.Deployment = "" },
			want:   "AUTOSCALER_EVENT_PROCESSOR_DEPLOYMENT must not be empty",
		},
		{
			name:   "empty queue",
			mutate: func(target *AutoscalerTarget) { target.Queue = "" },
			want:   "AUTOSCALER_EVENT_PROCESSOR_QUEUE must not be empty",
		},
		{
			name:   "zero min replicas",
			mutate: func(target *AutoscalerTarget) { target.MinReplicas = 0 },
			want:   "AUTOSCALER_EVENT_PROCESSOR_MIN_REPLICAS must be >= 1",
		},
		{
			name:   "max replicas below min",
			mutate: func(target *AutoscalerTarget) { target.MaxReplicas = target.MinReplicas - 1 },
			want:   "AUTOSCALER_EVENT_PROCESSOR_MAX_REPLICAS must be >= MIN_REPLICAS",
		},
		{
			name:   "max replicas above int32",
			mutate: func(target *AutoscalerTarget) { target.MaxReplicas = math.MaxInt32 + 1 },
			want:   "AUTOSCALER_EVENT_PROCESSOR_MAX_REPLICAS must be <= 2147483647",
		},
		{
			name:   "zero scale up window",
			mutate: func(target *AutoscalerTarget) { target.ScaleUpWindow = 0 },
			want:   "AUTOSCALER_EVENT_PROCESSOR_SCALE_UP_WINDOW must be >= 1",
		},
		{
			name:   "zero scale down window",
			mutate: func(target *AutoscalerTarget) { target.ScaleDownWindow = 0 },
			want:   "AUTOSCALER_EVENT_PROCESSOR_SCALE_DOWN_WINDOW must be >= 1",
		},
		{
			name: "negative scale up backlog",
			mutate: func(target *AutoscalerTarget) {
				target.ScaleUpBacklog = -1
				target.ScaleDownBacklog = -2
			},
			want: "AUTOSCALER_EVENT_PROCESSOR_SCALE_UP_BACKLOG must be >= 0",
		},
		{
			name:   "zero scale down backlog",
			mutate: func(target *AutoscalerTarget) { target.ScaleDownBacklog = 0 },
			want:   "AUTOSCALER_EVENT_PROCESSOR_SCALE_DOWN_BACKLOG must be >= 1",
		},
		{
			name:   "scale down backlog above scale up backlog",
			mutate: func(target *AutoscalerTarget) { target.ScaleDownBacklog = target.ScaleUpBacklog + 1 },
			want:   "AUTOSCALER_EVENT_PROCESSOR_SCALE_DOWN_BACKLOG must be < SCALE_UP_BACKLOG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newConfig(t)
			tt.mutate(c.AutoscalerEventProcessor)

			err := c.ValidateAutoscaler()
			require.ErrorContains(t, err, tt.want)
			require.NotContains(t, err.Error(), historyAPIEnvPrefix, "the untouched target must stay valid")
		})
	}
}

func TestRegisterAutoscalerTargetDistinctFlags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	eventProcessor := registerAutoscalerTarget(fs, eventProcessorEnvPrefix, eventProcessorFlagPrefix, eventProcessorDefaults())
	historyAPI := registerAutoscalerTarget(fs, historyAPIEnvPrefix, historyAPIFlagPrefix, historyAPIDefaults())

	names := map[string]bool{}
	fs.VisitAll(func(f *flag.Flag) { names[f.Name] = true })
	require.Len(t, names, 18, "both targets must register 9 flags each under distinct names")
	require.True(t, names["autoscalerEventProcessorMinReplicas"])
	require.True(t, names["autoscalerHistoryAPIMinReplicas"])

	err := fs.Parse([]string{"-autoscalerEventProcessorMinReplicas=3", "-autoscalerHistoryAPIMinReplicas=5"})
	require.NoError(t, err)
	require.Equal(t, 3, eventProcessor.MinReplicas)
	require.Equal(t, 5, historyAPI.MinReplicas)
}

func TestRegisterAutoscalerTargetEnvDefaults(t *testing.T) {
	t.Setenv("AUTOSCALER_EVENT_PROCESSOR_ENABLED", "false")
	t.Setenv("AUTOSCALER_EVENT_PROCESSOR_DEPLOYMENT", "custom-processor")
	t.Setenv("AUTOSCALER_EVENT_PROCESSOR_QUEUE", "custom_queue")
	t.Setenv("AUTOSCALER_EVENT_PROCESSOR_MIN_REPLICAS", "3")
	t.Setenv("AUTOSCALER_EVENT_PROCESSOR_MAX_REPLICAS", "42")
	t.Setenv("AUTOSCALER_EVENT_PROCESSOR_SCALE_UP_BACKLOG", "5000")
	t.Setenv("AUTOSCALER_EVENT_PROCESSOR_SCALE_UP_WINDOW", "6")
	t.Setenv("AUTOSCALER_EVENT_PROCESSOR_SCALE_DOWN_BACKLOG", "7")
	t.Setenv("AUTOSCALER_EVENT_PROCESSOR_SCALE_DOWN_WINDOW", "8")

	eventProcessor, historyAPI := newTargets(t)

	require.Equal(t, AutoscalerTarget{
		Name:             "event-processor",
		envPrefix:        eventProcessorEnvPrefix,
		Enabled:          false,
		Deployment:       "custom-processor",
		Queue:            "custom_queue",
		MinReplicas:      3,
		MaxReplicas:      42,
		ScaleUpBacklog:   5000,
		ScaleUpWindow:    6,
		ScaleDownBacklog: 7,
		ScaleDownWindow:  8,
	}, *eventProcessor)

	// the other prefix keeps its own defaults
	wantHistoryAPI := historyAPIDefaults()
	wantHistoryAPI.envPrefix = historyAPIEnvPrefix
	require.Equal(t, wantHistoryAPI, *historyAPI)
}

func TestAutoscalerTargets(t *testing.T) {
	c := newConfig(t)

	targets := c.AutoscalerTargets()
	require.Len(t, targets, 2)
	require.Equal(t, "event-processor", targets[0].Name)
	require.Equal(t, "history-api", targets[1].Name)
}
