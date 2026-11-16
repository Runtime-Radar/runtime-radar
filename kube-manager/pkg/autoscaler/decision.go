package autoscaler

import (
	"fmt"
	"time"
)

// Observation is what a tick knows about a target.
type Observation struct {
	Backlog            int
	Replicas           int32     // spec.replicas
	Available          int32     // status.availableReplicas
	Generation         int64     // metadata.generation
	ObservedGeneration int64     // status.observedGeneration
	LastScaled         time.Time // zero when never scaled
}

// Decision is either an action (From -> Desired, explained by Reason) or a Skip reason.
type Decision struct {
	From    int32
	Desired int32
	Reason  string
	Skip    string
}

func (d Decision) String() string {
	if d.Skip != "" {
		return "skip: " + d.Skip
	}

	return fmt.Sprintf("%d -> %d (%s)", d.From, d.Desired, d.Reason)
}

// decide advances the counters with the observed backlog and reports what to do. The caller
// resets the counter of a fired decision once the action succeeded.
//
// An unhealthy or mid-rollout deployment only holds back actions that add replicas: piling more
// onto a deployment that cannot start what it has is pointless, while shedding them is still safe.
func decide(cfg TargetConfig, st *targetState, obs Observation, now time.Time, cooldown time.Duration) Decision {
	switch {
	case obs.Backlog > cfg.UpBacklog:
		st.high++
		st.resetLow()
	case obs.Backlog < cfg.DownBacklog:
		st.low++
		st.resetHigh()
	default:
		st.reset()
	}

	// Counters keep counting during a hold, so a window filled during a rollout fires right after it.
	var hold string

	switch {
	case obs.Available < obs.Replicas:
		hold = "unhealthy"
	case obs.ObservedGeneration < obs.Generation:
		hold = "rollout"
	}

	// A future annotation is clock skew and counts as elapsed, so the cooldown never stretches by it.
	if since := now.Sub(obs.LastScaled); !obs.LastScaled.IsZero() && since >= 0 && since < cooldown {
		return Decision{Skip: "cooldown"}
	}

	// Bounds win over windows; this also undoes a Helm upgrade resetting spec.replicas.
	switch {
	case obs.Replicas < cfg.Min:
		if hold != "" {
			return Decision{Skip: hold}
		}

		return Decision{
			From:    obs.Replicas,
			Desired: cfg.Min,
			Reason:  fmt.Sprintf("replicas %d below min %d", obs.Replicas, cfg.Min),
		}
	case obs.Replicas > cfg.Max:
		return Decision{
			From:    obs.Replicas,
			Desired: cfg.Max,
			Reason:  fmt.Sprintf("replicas %d above max %d", obs.Replicas, cfg.Max),
		}
	}

	switch {
	case st.high >= cfg.UpWindow:
		if hold != "" {
			return Decision{Skip: hold}
		}

		if obs.Replicas >= cfg.Max {
			return Decision{Skip: "at max"}
		}

		return Decision{
			From:    obs.Replicas,
			Desired: obs.Replicas + 1,
			Reason:  fmt.Sprintf("backlog %d above %d for %d samples", obs.Backlog, cfg.UpBacklog, st.high),
		}
	case st.low >= cfg.DownWindow:
		if obs.Replicas <= cfg.Min {
			return Decision{Skip: "at min"}
		}

		return Decision{
			From:    obs.Replicas,
			Desired: obs.Replicas - 1,
			Reason:  fmt.Sprintf("backlog %d below %d for %d samples", obs.Backlog, cfg.DownBacklog, st.low),
		}
	case st.high == 0 && st.low == 0:
		return Decision{Skip: "neutral"}
	}

	return Decision{Skip: "window"}
}
