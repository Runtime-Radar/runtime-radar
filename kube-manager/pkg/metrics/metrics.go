package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/runtime-radar/runtime-radar/lib/metrics"
)

const (
	K8SEntity = "k8s_entity"

	EventType = "event_type"
	Add       = "add"
	Update    = "update"
	Delete    = "delete"
)

var (
	commonMetrics []prometheus.Collector

	InformerEventsCount = addToCommon(prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "informer_events_count",
		Help: "The number of successfully processed informer events",
	}, []string{K8SEntity, EventType}))
	InformerErrorsCount = addToCommon(prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "informer_errors_count",
		Help: "The number of errors that occurred while attempting to sync with Kubernetes API",
	}, []string{K8SEntity}))

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
