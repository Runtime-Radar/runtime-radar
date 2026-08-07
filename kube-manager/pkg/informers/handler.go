package informers

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/metrics"
	"k8s.io/client-go/tools/cache"
)

func (i *Informer[T]) add(obj interface{}) {
	entity, ok := obj.(Object)
	if !ok {
		log.Error().Err(ErrUnableCast).Msg("Can't add object to informer")
		return
	}

	metrics.InformerEventsCount.With(prometheus.Labels{metrics.K8SEntity: i.informerType, metrics.EventType: metrics.Add}).Inc()
	log.Debug().Msgf("Add %s %s/%s", i.informerType, entity.GetNamespace(), entity.GetName())
}

func (i *Informer[T]) update(_, newObj interface{}) {
	entity, ok := newObj.(Object)
	if !ok {
		log.Error().Err(ErrUnableCast).Msg("Can't update object in informer")
		return
	}

	metrics.InformerEventsCount.With(prometheus.Labels{metrics.K8SEntity: i.informerType, metrics.EventType: metrics.Update}).Inc()
	log.Debug().Msgf("Update %s %s/%s", i.informerType, entity.GetNamespace(), entity.GetName())
}

func (i *Informer[T]) delete(obj interface{}) {
	switch entity := obj.(type) {
	case Object:
		log.Debug().Msgf("Delete %s %s/%s", i.informerType, entity.GetNamespace(), entity.GetName())
	case cache.DeletedFinalStateUnknown:
		log.Debug().Msgf("Delete %s, DeletedFinalStateUnknown detected", i.informerType)
	default:
		log.Error().Msgf("Can't delete object from informer, unknown type: %T", obj)
	}

	metrics.InformerEventsCount.With(prometheus.Labels{metrics.K8SEntity: i.informerType, metrics.EventType: metrics.Delete}).Inc()
}

func (i *Informer[T]) errorHandler(_ *cache.Reflector, err error) {
	metrics.InformerErrorsCount.With(prometheus.Labels{metrics.K8SEntity: i.informerType}).Inc()
	log.Error().Err(err).Str("informer", i.informerType).Msg("Informer error")
}
