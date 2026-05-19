package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/runtime-radar/runtime-radar/lib/metrics"
)

const (
	Code   = "code"
	Method = "method"
	Path   = "path"
)

var (
	commonMetrics []prometheus.Collector

	RequestCounter = addToCommon(prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{Method, Path, Code},
	))
	RequestDurationHistogram = addToCommon(prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request durations.",
			Buckets: prometheus.DefBuckets,
		},
	))

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

func addToCommon[T prometheus.Collector](metric T) T {
	commonMetrics = append(commonMetrics, metric)
	return metric
}
