package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	lib_metrics "github.com/runtime-radar/runtime-radar/lib/metrics"
)

var (
	commonMetrics []prometheus.Collector
)

// PrepareRegistry creates and registers service global metrics.
func PrepareRegistry(service, cluster string, m ...prometheus.Collector) (*lib_metrics.Registry, error) {
	var mm []prometheus.Collector
	mm = append(mm, commonMetrics...)
	mm = append(mm, m...)

	return lib_metrics.NewRegistry(service, cluster, mm...)
}

func addToCommon[T prometheus.Collector](metric T) T {
	commonMetrics = append(commonMetrics, metric)
	return metric
}
