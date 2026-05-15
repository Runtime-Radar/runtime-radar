package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/runtime-radar/runtime-radar/lib/metrics"
)

const HasError = "hasError"

var (
	commonMetrics []prometheus.Collector

	ClusterStateGauge = addToCommon(prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cluster_state_gauge",
		Help: "Current cluster state (0-Central, 1-ChildUnregistered, 2-ChildRegistered)",
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
