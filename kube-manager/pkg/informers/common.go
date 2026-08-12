package informers

import (
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

var (
	ErrNotFound        = errors.New("object not found")
	ErrUnableCast      = errors.New("unable to cast")
	ErrUnableToConvert = errors.New("unable to convert to object")
	ErrAlreadyAdded    = errors.New("informer for namespace is already added")
)

type Object interface {
	runtime.Object
	metav1.Object
}

type Setter interface {
	AddInformer(ns string, informer cache.SharedIndexInformer) error
	RemoveInformer(ns string) error
	Type() string
}

type Getter[T Object] interface {
	Get(namespace, name string) (T, error)

	// List filtered by namespaces. If the namespace list is empty,
	// all entities without a namespace filter are used
	List(namespaces ...string) ([]T, int)
}
