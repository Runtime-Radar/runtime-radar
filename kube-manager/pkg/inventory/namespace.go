package inventory

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/informers"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/metrics"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

func (i *Inventory) Get(_, name string) (*corev1.Namespace, error) {
	item, exists, err := i.nsInformer.GetStore().GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, informers.ErrNotFound
	}

	return item.(*corev1.Namespace), nil
}

func (i *Inventory) List(_ ...string) ([]*corev1.Namespace, int) {
	l := i.nsInformer.GetStore().List()
	nss := make([]*corev1.Namespace, 0, len(l))
	for _, n := range l {
		nss = append(nss, n.(*corev1.Namespace))
	}

	return nss, len(nss)
}

func (i *Inventory) addNamespaceInformer() (cache.SharedIndexInformer, error) {
	nsi := i.nsFactory.InformerFor(&corev1.Namespace{}, func(k kubernetes.Interface, duration time.Duration) cache.SharedIndexInformer {
		return v1.NewNamespaceInformer(k, duration, cache.Indexers{})
	})

	_, err := nsi.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    i.addNamespaceHandler,
		UpdateFunc: i.updateNamespaceHandler,
		DeleteFunc: i.deleteNamespaceHandler,
	})
	if err != nil {
		return nil, fmt.Errorf("can't add event handler for namespace informer: %w", err)
	}

	err = nsi.SetWatchErrorHandler(func(_ *cache.Reflector, err error) {
		metrics.InformerErrorsCount.With(prometheus.Labels{metrics.K8SEntity: "namespace"}).Inc()
		log.Warn().Msgf("Informer for type '%s' failed with an error: %v", "namespace", err)
	})
	if err != nil {
		return nil, fmt.Errorf("can't add error handler for namespace informer: %w", err)
	}

	return nsi, nil
}

func (i *Inventory) addNamespaceHandler(obj interface{}) {
	entity, ok := obj.(*corev1.Namespace)
	if !ok {
		log.Error().Msg(fmt.Sprintf("entity is not of type '*corev1.Namespace', current type: '%T'", obj))
		return
	}

	metrics.InformerEventsCount.With(prometheus.Labels{metrics.K8SEntity: "namespace", metrics.EventType: metrics.Add}).Inc()
	log.Debug().Msgf("Add 'namespace' %s", entity.GetName())

	if _, started := i.dynamicFactories.get(entity.GetName()); started {
		return
	}

	allow, deny := i.updater.Maps()
	if len(allow) > 0 {
		if _, isAllowed := allow[entity.GetName()]; !isAllowed {
			return
		}

		err := i.addFactory(entity.GetName())
		if err != nil {
			log.Error().Err(err).Msgf("can't add factory for namespace %s", entity.GetName())
		}

		return
	}

	if _, isDeny := deny[entity.GetName()]; !isDeny {
		err := i.addFactory(entity.GetName())
		if err != nil {
			log.Error().Err(err).Msgf("can't add factory for namespace %s", entity.GetName())
		}
	}
}

func (i *Inventory) updateNamespaceHandler(_, newObj interface{}) {
	entity, ok := newObj.(metav1.Object)
	if !ok {
		return
	}

	metrics.InformerEventsCount.With(prometheus.Labels{metrics.K8SEntity: "namespace", metrics.EventType: metrics.Update}).Inc()
	log.Debug().Msgf("Update 'namespace' %s", entity.GetName())
}

func (i *Inventory) deleteNamespaceHandler(obj interface{}) {
	entity, ok := obj.(*corev1.Namespace)
	if !ok {
		log.Error().Msg(fmt.Sprintf("entity %T is not of type *corev1.Namespace", obj))
		return
	}

	metrics.InformerEventsCount.With(prometheus.Labels{metrics.K8SEntity: "namespace", metrics.EventType: metrics.Delete}).Inc()
	log.Debug().Msgf("Delete 'namespace' %s", entity.GetName())

	if _, started := i.dynamicFactories.get(entity.GetName()); !started {
		return
	}

	if err := i.removeFactory(entity.GetName()); err != nil {
		log.Error().Err(err).Msgf("can't remove factory for namespace %s", entity.GetName())
	}
}
