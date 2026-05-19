package publisher

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/lib/rabbit"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/metrics"
	"github.com/runtime-radar/runtime-radar/runtime-monitor/pkg/monitor"
)

type Publisher struct {
	Monitor         monitor.Monitor
	PublishConsumer rabbit.PublishConsumer
}

func (p *Publisher) Run(stop <-chan struct{}) {
	log.Info().Msgf("Events publisher started")
	defer log.Info().Msgf("Events publisher stopped")

	events := p.Monitor.Events()
	ctx := context.Background()

	for {
		select {
		case ev := <-events:
			if err := p.PublishConsumer.Publish(ctx, ev); err != nil {
				log.Error().Err(err).Msgf("Can't publish event")
				metrics.RabbitAddEventsFailureCount.Inc()

				continue
			}
			metrics.RabbitAddEventsSuccessCount.Inc()

		case <-stop:
			return
		}
	}
}
