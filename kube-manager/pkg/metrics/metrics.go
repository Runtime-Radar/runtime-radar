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

	Target    = "target"
	Queue     = "queue"
	Direction = "direction"
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

	AutoscalerQueueBacklog = addToCommon(prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "autoscaler_queue_backlog",
		Help: "The number of messages ready for delivery in the queue of an autoscaler target",
	}, []string{Target, Queue}))
	AutoscalerQueueConsumers = addToCommon(prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "autoscaler_queue_consumers",
		Help: "The number of consumers attached to the queue of an autoscaler target",
	}, []string{Target, Queue}))
	AutoscalerReplicas = addToCommon(prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "autoscaler_replicas",
		Help: "The number of replicas of a target deployment observed at the last sample",
	}, []string{Target}))
	AutoscalerScaleActionsCount = addToCommon(prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "autoscaler_scale_actions_count",
		Help: "The number of scaling actions the autoscaler performed",
	}, []string{Target, Direction}))
	AutoscalerIsLeader = addToCommon(prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "autoscaler_is_leader",
		Help: "Whether this instance holds the autoscaler leader lease",
	}))

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
