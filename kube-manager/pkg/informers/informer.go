package informers

import (
	"fmt"
	"sync"

	"k8s.io/client-go/tools/cache"
)

type Informer[T Object] struct {
	informers    map[string]cache.SharedIndexInformer
	informerType string

	m sync.RWMutex
}

// New creates an aggregator informer object for each K8S type for which data needs to be collected.
// In it, we store a SharedIndexInformer map for each namespace that needs to be listened to.
// We connect SharedIndexInformer using a dynamic factory created for each namespace.
func New[T Object](informerType string) *Informer[T] {
	return &Informer[T]{
		informers:    make(map[string]cache.SharedIndexInformer),
		informerType: informerType,
	}
}

// AddInformer adds an informer of a certain type for a given namespace
func (i *Informer[T]) AddInformer(ns string, informer cache.SharedIndexInformer) error {
	if _, ok := i.getInformer(ns); ok {
		return ErrAlreadyAdded
	}

	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    i.add,
		UpdateFunc: i.update,
		DeleteFunc: i.delete,
	})
	if err != nil {
		return fmt.Errorf("can't add event handler: %w", err)
	}

	err = informer.SetWatchErrorHandler(i.errorHandler)
	if err != nil {
		return fmt.Errorf("can't set error handler: %w", err)
	}

	i.m.Lock()
	defer i.m.Unlock()

	i.informers[ns] = informer

	return nil
}

// RemoveInformer removes an informer of a specific type for a namespace that no longer needs to be listened to
func (i *Informer[T]) RemoveInformer(ns string) error {
	i.m.Lock()
	defer i.m.Unlock()

	delete(i.informers, ns)

	return nil
}

func (i *Informer[T]) getInformer(namespace string) (cache.SharedIndexInformer, bool) {
	i.m.RLock()
	defer i.m.RUnlock()

	inf, ok := i.informers[namespace]
	if !ok || !inf.HasSynced() {
		return nil, false
	}

	return inf, ok
}

func (i *Informer[T]) empty() T {
	t := new(T)
	return *t
}

func (i *Informer[T]) Type() string {
	return i.informerType
}
