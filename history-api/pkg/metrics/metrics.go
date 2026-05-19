package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/runtime-radar/runtime-radar/lib/metrics"
)

const (
	queueLabel      = "queue"
	isConsumerLabel = "is_consumer"
)

var (
	commonMetrics []prometheus.Collector

	RabbitBrokerConnectionStateGauge = addToCommon(prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rabbit_broker_connection_state_gauge",
		Help: "Availability of broker connection to rabbit (1 - available, 0 - not available)",
	}, []string{queueLabel, isConsumerLabel}))

	// Example addToCommon definition:
	//
	//	ProxySuccessCount = addToCommon(prometheus.NewCounterVec(prometheus.CounterOpts{
	//		Name: "proxy_success_count",
	//		Help: "The number of successfully processed messages on proxy service",
	//	}, []string{myLabel}))
)

// PrepareRegistry creates and registers service global metrics
func PrepareRegistry(service, cluster string, m ...prometheus.Collector) (*metrics.Registry, error) {
	var mm []prometheus.Collector
	mm = append(mm, commonMetrics...)
	mm = append(mm, m...)

	return metrics.NewRegistry(service, cluster, mm...)
}

func RabbitStateReporter(queue string, isConsumer bool) func(bool) {
	return func(isAlive bool) {
		var value float64
		if isAlive {
			value = 1
		}

		RabbitBrokerConnectionStateGauge.With(prometheus.Labels{queueLabel: queue, isConsumerLabel: strconv.FormatBool(isConsumer)}).Set(value)
	}
}

func addToCommon[T prometheus.Collector](metric T) T {
	commonMetrics = append(commonMetrics, metric)
	return metric
}
