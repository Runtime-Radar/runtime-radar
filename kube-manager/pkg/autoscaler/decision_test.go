package autoscaler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testCooldown = 90 * time.Second

var testNow = time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)

func testTarget() TargetConfig {
	return TargetConfig{
		Name:        "event-processor",
		Deployment:  "event-processor",
		Queue:       "runtime_events",
		Min:         2,
		Max:         10,
		UpBacklog:   1000,
		UpWindow:    4,
		DownBacklog: 10,
		DownWindow:  40,
	}
}

// pinnedTarget has min == max, so no window action is possible.
func pinnedTarget(replicas int32) TargetConfig {
	cfg := testTarget()
	cfg.Min, cfg.Max = replicas, replicas

	return cfg
}

func healthy(backlog int, replicas int32) Observation {
	return Observation{
		Backlog:            backlog,
		Replicas:           replicas,
		Available:          replicas,
		Generation:         3,
		ObservedGeneration: 3,
	}
}

func TestDecide(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cfg      TargetConfig
		state    targetState
		obs      Observation
		want     Decision
		wantHigh int
		wantLow  int
	}{
		{
			name:  "neutral band resets counters",
			cfg:   testTarget(),
			state: targetState{high: 3, low: 5},
			obs:   healthy(500, 3),
			want:  Decision{Skip: "neutral"},
		},
		{
			name:     "high backlog one sample short of the window",
			cfg:      testTarget(),
			state:    targetState{high: 2},
			obs:      healthy(1432, 2),
			want:     Decision{Skip: "window"},
			wantHigh: 3,
		},
		{
			name:     "high backlog for exactly the window scales up",
			cfg:      testTarget(),
			state:    targetState{high: 3},
			obs:      healthy(1432, 2),
			want:     Decision{From: 2, Desired: 3, Reason: "backlog 1432 above 1000 for 4 samples"},
			wantHigh: 4,
		},
		{
			name:     "high backlog at max replicas",
			cfg:      testTarget(),
			state:    targetState{high: 3},
			obs:      healthy(1432, 10),
			want:     Decision{Skip: "at max"},
			wantHigh: 4,
		},
		{
			name:    "low backlog for the window scales down",
			cfg:     testTarget(),
			state:   targetState{low: 39},
			obs:     healthy(3, 3),
			want:    Decision{From: 3, Desired: 2, Reason: "backlog 3 below 10 for 40 samples"},
			wantLow: 40,
		},
		{
			name:    "low backlog at min replicas",
			cfg:     testTarget(),
			state:   targetState{low: 39},
			obs:     healthy(3, 2),
			want:    Decision{Skip: "at min"},
			wantLow: 40,
		},
		{
			name:  "counters advance during cooldown",
			cfg:   testTarget(),
			state: targetState{high: 3},
			obs: Observation{
				Backlog: 1432, Replicas: 2, Available: 2,
				Generation: 3, ObservedGeneration: 3,
				LastScaled: testNow.Add(-30 * time.Second),
			},
			want:     Decision{Skip: "cooldown"},
			wantHigh: 4,
		},
		{
			name:  "elapsed cooldown does not block",
			cfg:   testTarget(),
			state: targetState{high: 3},
			obs: Observation{
				Backlog: 1432, Replicas: 2, Available: 2,
				Generation: 3, ObservedGeneration: 3,
				LastScaled: testNow.Add(-testCooldown),
			},
			want:     Decision{From: 2, Desired: 3, Reason: "backlog 1432 above 1000 for 4 samples"},
			wantHigh: 4,
		},
		{
			name:     "zero last scaled means no cooldown",
			cfg:      testTarget(),
			state:    targetState{high: 3},
			obs:      healthy(1432, 2),
			want:     Decision{From: 2, Desired: 3, Reason: "backlog 1432 above 1000 for 4 samples"},
			wantHigh: 4,
		},
		{
			name:  "unavailable replicas hold the scale-up",
			cfg:   testTarget(),
			state: targetState{high: 3},
			obs: Observation{
				Backlog: 1432, Replicas: 2, Available: 1,
				Generation: 3, ObservedGeneration: 3,
			},
			want:     Decision{Skip: "unhealthy"},
			wantHigh: 4,
		},
		{
			name:  "rollout in progress holds the scale-up",
			cfg:   testTarget(),
			state: targetState{high: 3},
			obs: Observation{
				Backlog: 1432, Replicas: 2, Available: 2,
				Generation: 3, ObservedGeneration: 2,
			},
			want:     Decision{Skip: "rollout"},
			wantHigh: 4,
		},
		{
			name:  "unavailable replicas do not hold the scale-down",
			cfg:   testTarget(),
			state: targetState{low: 39},
			obs: Observation{
				Backlog: 3, Replicas: 3, Available: 2,
				Generation: 3, ObservedGeneration: 3,
			},
			want:    Decision{From: 3, Desired: 2, Reason: "backlog 3 below 10 for 40 samples"},
			wantLow: 40,
		},
		{
			name:  "unavailable replicas do not hold the clamp to max",
			cfg:   testTarget(),
			state: targetState{},
			obs: Observation{
				Backlog: 500, Replicas: 12, Available: 8,
				Generation: 3, ObservedGeneration: 3,
			},
			want: Decision{From: 12, Desired: 10, Reason: "replicas 12 above max 10"},
		},
		{
			name:  "unavailable replicas hold the clamp to min",
			cfg:   testTarget(),
			state: targetState{},
			obs: Observation{
				Backlog: 500, Replicas: 1, Available: 0,
				Generation: 3, ObservedGeneration: 3,
			},
			want: Decision{Skip: "unhealthy"},
		},
		{
			name:  "replicas below min are raised without a window",
			cfg:   testTarget(),
			state: targetState{},
			obs:   healthy(500, 1),
			want:  Decision{From: 1, Desired: 2, Reason: "replicas 1 below min 2"},
		},
		{
			name:  "replicas above max are lowered without a window",
			cfg:   testTarget(),
			state: targetState{},
			obs:   healthy(500, 12),
			want:  Decision{From: 12, Desired: 10, Reason: "replicas 12 above max 10"},
		},
		{
			name:  "backlog exactly at the up threshold is neutral",
			cfg:   testTarget(),
			state: targetState{high: 3},
			obs:   healthy(1000, 2),
			want:  Decision{Skip: "neutral"},
		},
		{
			name:  "backlog exactly at the down threshold is neutral",
			cfg:   testTarget(),
			state: targetState{low: 39},
			obs:   healthy(10, 3),
			want:  Decision{Skip: "neutral"},
		},
		{
			name:     "min equal to max leaves the up window nothing to do",
			cfg:      pinnedTarget(3),
			state:    targetState{high: 3},
			obs:      healthy(1432, 3),
			want:     Decision{Skip: "at max"},
			wantHigh: 4,
		},
		{
			name:    "min equal to max leaves the down window nothing to do",
			cfg:     pinnedTarget(3),
			state:   targetState{low: 39},
			obs:     healthy(3, 3),
			want:    Decision{Skip: "at min"},
			wantLow: 40,
		},
		{
			name:  "annotation in the future does not extend the cooldown",
			cfg:   testTarget(),
			state: targetState{high: 3},
			obs: Observation{
				Backlog: 1432, Replicas: 2, Available: 2,
				Generation: 3, ObservedGeneration: 3,
				LastScaled: testNow.Add(time.Hour),
			},
			want:     Decision{From: 2, Desired: 3, Reason: "backlog 1432 above 1000 for 4 samples"},
			wantHigh: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			st := tt.state
			got := decide(tt.cfg, &st, tt.obs, testNow, testCooldown)

			require.Equal(t, tt.want, got)
			require.Equal(t, targetState{high: tt.wantHigh, low: tt.wantLow}, st)
		})
	}
}

func TestDecisionString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		d    Decision
		want string
	}{
		{
			name: "scale up",
			d:    Decision{From: 2, Desired: 3, Reason: "backlog 1432 above 1000 for 4 samples"},
			want: "2 -> 3 (backlog 1432 above 1000 for 4 samples)",
		},
		{
			name: "scale down",
			d:    Decision{From: 3, Desired: 2, Reason: "backlog 3 below 10 for 40 samples"},
			want: "3 -> 2 (backlog 3 below 10 for 40 samples)",
		},
		{
			name: "skip",
			d:    Decision{Skip: "cooldown"},
			want: "skip: cooldown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, tt.d.String())
		})
	}
}
