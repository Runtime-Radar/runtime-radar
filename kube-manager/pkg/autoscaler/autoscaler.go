// Package autoscaler scales Deployments from the backlog of the RabbitMQ queue they consume.
package autoscaler

import (
	"context"
	"errors"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/metrics"
)

const (
	directionUp   = "up"
	directionDown = "down"
)

// Autoscaler scales the target Deployments from the backlog of the queues they consume.
type Autoscaler struct {
	Targets  []TargetConfig
	Scaler   *Scaler
	Interval time.Duration
	Cooldown time.Duration
	Clock    func() time.Time // defaults to time.Now

	states []targetState // per-target counters, parallel to Targets
}

// Run ticks until ctx is cancelled. Counters start empty: a leadership term that ended mid-window
// must not carry samples into a term nobody watched the backlog in between.
func (a *Autoscaler) Run(ctx context.Context) {
	a.states = make([]targetState, len(a.Targets))

	// The per-target gauges describe what this replica sees while it leads; an ex-leader exporting
	// its last sample forever would read as a live one.
	defer a.forgetGauges()

	log.Info().Dur("interval", a.Interval).Dur("cooldown", a.Cooldown).Msg("Autoscaler started")

	ticker := time.NewTicker(a.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Autoscaler stopped")

			return

		case <-ticker.C:
			a.tick(ctx)
		}
	}
}

func (a *Autoscaler) tick(ctx context.Context) {
	for i, cfg := range a.Targets {
		a.tickTarget(ctx, cfg, &a.states[i])
	}
}

func (a *Autoscaler) tickTarget(ctx context.Context, cfg TargetConfig, st *targetState) {
	logger := log.With().
		Str("target", cfg.Name).
		Str("queue", cfg.Queue).
		Str("deployment", cfg.Deployment).
		Logger()

	backlog, consumers, err := cfg.Prober.InspectQueue(ctx)
	if err != nil {
		// The gauges keep their last values: a broker outage is not an empty queue.
		st.reset()
		logger.Warn().Err(err).Msg("Can't probe queue")

		return
	}

	backlogGauge(cfg).Set(float64(backlog))
	consumersGauge(cfg).Set(float64(consumers))

	// The broker declares the queue before inspecting it, so a queue name that matches no consumer
	// reads as a permanently empty one instead of failing.
	if consumers == 0 {
		logger.Warn().Msg("Nothing consumes the probed queue, check that its name is the one the deployment reads")
	}

	dep, err := a.Scaler.Get(ctx, cfg.Deployment)
	if err != nil {
		logger.Error().Err(err).Msg("Can't read deployment")

		return
	}

	obs := ObservationFrom(dep, backlog)
	replicasGauge(cfg).Set(float64(obs.Replicas))

	decision := decide(cfg, st, obs, a.now(), a.Cooldown)
	if decision.Skip != "" {
		logger.Debug().Int("backlog", backlog).Int32("from", obs.Replicas).Msg(decision.String())

		return
	}

	if err = a.Scaler.Apply(ctx, dep, decision.Desired, a.now(), decision.Reason); err != nil {
		if errors.Is(err, ErrConflict) {
			logger.Info().Msg("Deployment changed since it was read, skipping the action")

			return
		}

		logger.Error().Err(err).Msg("Can't scale deployment")

		return
	}

	direction := directionDown

	if decision.Desired > decision.From {
		direction = directionUp
		st.resetHigh()
	} else {
		st.resetLow()
	}

	metrics.AutoscalerScaleActionsCount.With(prometheus.Labels{
		metrics.Target:    cfg.Name,
		metrics.Direction: direction,
	}).Inc()

	replicasGauge(cfg).Set(float64(decision.Desired))

	logger.Info().
		Int("backlog", backlog).
		Int32("from", decision.From).
		Int32("to", decision.Desired).
		Str("reason", decision.Reason).
		Msg(decision.String())
}

func (a *Autoscaler) now() time.Time {
	if a.Clock == nil {
		return time.Now()
	}

	return a.Clock()
}

func (a *Autoscaler) forgetGauges() {
	for _, cfg := range a.Targets {
		metrics.AutoscalerQueueBacklog.Delete(queueLabels(cfg))
		metrics.AutoscalerQueueConsumers.Delete(queueLabels(cfg))
		metrics.AutoscalerReplicas.Delete(targetLabels(cfg))
	}
}

func backlogGauge(cfg TargetConfig) prometheus.Gauge {
	return metrics.AutoscalerQueueBacklog.With(queueLabels(cfg))
}

func consumersGauge(cfg TargetConfig) prometheus.Gauge {
	return metrics.AutoscalerQueueConsumers.With(queueLabels(cfg))
}

func replicasGauge(cfg TargetConfig) prometheus.Gauge {
	return metrics.AutoscalerReplicas.With(targetLabels(cfg))
}

func queueLabels(cfg TargetConfig) prometheus.Labels {
	return prometheus.Labels{metrics.Target: cfg.Name, metrics.Queue: cfg.Queue}
}

func targetLabels(cfg TargetConfig) prometheus.Labels {
	return prometheus.Labels{metrics.Target: cfg.Name}
}
