package inventory

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/informers"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
)

// Default namespace for entities that do not have a namespace
const defaultNamespace = ""

var nonNamespaceEntities = map[string]struct{}{
	"nodes": {},
}

func (i *Inventory) addFactory(ns string) error {
	f := dynamicinformer.NewFilteredDynamicSharedInformerFactory(i.dynamicClient, i.syncInterval, ns, nil)

	stopCh := make(chan struct{})

	err := i.dynamicFactories.add(ns, DynamicInformerFactory{
		factory: f,
		stopCh:  stopCh,
	})

	switch {
	case errors.Is(err, ErrFactoryExists):
		return nil
	case err != nil:
		return err
	}
	gvrs, err := i.resourcer.Get(i.staticClient.Discovery())
	if err != nil {
		return fmt.Errorf("can't get server resources: %w", err)
	}
	gvrMap := make(map[string]schema.GroupVersionResource, len(gvrs))
	for _, gvr := range gvrs {
		gvrMap[gvr.Resource] = gvr
	}

	for _, inf := range i.informers {
		_, isNonNamespaced := nonNamespaceEntities[inf.Type()]

		// if entity does not support namespaces, use one informer with defaultNamespace
		if (isNonNamespaced && ns != defaultNamespace) || (!isNonNamespaced && ns == defaultNamespace) {
			continue
		}

		gvr, ok := gvrMap[inf.Type()]
		if !ok {
			i.dynamicFactories.remove(ns)
			return fmt.Errorf("server resources for type '%s' not found", inf.Type())
		}

		informer := f.ForResource(gvr).Informer()
		log.Debug().Str("namespace", ns).Str("entity", inf.Type()).Msg("add dynamic informer")

		err = inf.AddInformer(ns, informer)
		switch {
		case errors.Is(err, informers.ErrAlreadyAdded):
			log.Debug().Str("namespace", ns).Str("entity", inf.Type()).Msg(informers.ErrAlreadyAdded.Error())
		case err != nil:
			return fmt.Errorf("can't add informer for type '%s': %w", inf.Type(), err)
		}
	}

	f.Start(stopCh)
	f.WaitForCacheSync(stopCh)

	return nil
}

func (i *Inventory) removeFactory(ns string) error {
	for _, inf := range i.informers {
		err := inf.RemoveInformer(ns)
		if err != nil {
			return fmt.Errorf("can't remove '%s' informer for type '%s': %w", ns, inf.Type(), err)
		}
	}

	i.dynamicFactories.remove(ns)

	return nil
}
