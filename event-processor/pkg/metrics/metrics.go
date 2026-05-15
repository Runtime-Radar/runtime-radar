package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/runtime-radar/runtime-radar/lib/metrics"
)

const (
	DetectorLabel   = "detector"
	queueLabel      = "queue"
	isConsumerLabel = "is_consumer"
)

var (
	commonMetrics []prometheus.Collector

	OperationTimeOneDetectorHistogram = addToCommon(prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "operation_time_one_detector_histogram",
		Help:    "Operation time of one detector",
		Buckets: []float64{.000005, .00001, .000025, .00005, .000075, .0001, .00025, .0005, .00075, .001, .0025, .005, .01, 0.05, .1, 1},
	}, []string{DetectorLabel},
	))
	OperationTimeAllDetectorsHistogram = addToCommon(prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "operation_time_all_detectors_histogram",
		Help:    "Operation time of the entire detector chain",
		Buckets: []float64{.000005, .00001, .000025, .00005, .000075, .0001, .00025, .0005, .00075, .001, .0025, .005, .01, 0.05, .1, .5, 1, 5, 10},
	}))
	DetectorThreatsCounter = addToCommon(prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "detector_threats_counter",
		Help: "Number of threats found by the detector",
	}, []string{DetectorLabel}))
	DetectorErrorsCounter = addToCommon(prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "detector_errors_counter",
		Help: "Number of detector calls that ended in errors",
	}, []string{DetectorLabel}))
	RabbitAddHistorySuccessCounter = addToCommon(prometheus.NewCounter(prometheus.CounterOpts{
		Name: "rabbit_add_history_success_count",
		Help: "Number of history events successfully added to the rabbitmq queue",
	}))
	RabbitAddHistoryFailureCounter = addToCommon(prometheus.NewCounter(prometheus.CounterOpts{
		Name: "rabbit_add_history_failure_count",
		Help: "Number of history events that failed to be added to the rabbitmq queue",
	}))
	DetectorsChainLength = addToCommon(prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "detectors_chain_length",
		Help:    "Length of detectors chain",
		Buckets: []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 50, 100},
	}))
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
