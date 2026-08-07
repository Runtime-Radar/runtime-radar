package inventory

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
	"k8s.io/client-go/dynamic/dynamicinformer"
)

var ErrFactoryExists = fmt.Errorf("factory already exists")

type FactoryCache struct {
	dynamicFactories map[string]DynamicInformerFactory

	m sync.RWMutex
}

type FactoryBuffer interface {
	get(ns string) (DynamicInformerFactory, bool)
	add(ns string, df DynamicInformerFactory) error
	remove(ns string)
	removeAll()
}

type DynamicInformerFactory struct {
	factory dynamicinformer.DynamicSharedInformerFactory
	stopCh  chan struct{}
}

func NewFactoryBuffer() *FactoryCache {
	return &FactoryCache{
		dynamicFactories: make(map[string]DynamicInformerFactory),
	}
}

func (f *FactoryCache) get(ns string) (DynamicInformerFactory, bool) {
	f.m.RLock()
	defer f.m.RUnlock()

	df, ok := f.dynamicFactories[ns]
	return df, ok
}

func (f *FactoryCache) add(ns string, df DynamicInformerFactory) error {
	f.m.Lock()
	defer f.m.Unlock()

	if _, ok := f.dynamicFactories[ns]; ok {
		return ErrFactoryExists
	}

	f.dynamicFactories[ns] = df
	return nil
}

func (f *FactoryCache) remove(ns string) {
	f.m.Lock()
	defer f.m.Unlock()

	df, ok := f.dynamicFactories[ns]
	if !ok {
		return
	}

	log.Debug().Str("namespace", ns).Msg("remove dynamic factory for namespace")

	close(df.stopCh)
	df.factory.Shutdown()
	delete(f.dynamicFactories, ns)
}

func (f *FactoryCache) removeAll() {
	f.m.Lock()
	defer f.m.Unlock()

	for _, df := range f.dynamicFactories {
		close(df.stopCh)
		df.factory.Shutdown()
	}

	f.dynamicFactories = map[string]DynamicInformerFactory{}
}
