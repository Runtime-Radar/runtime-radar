package informers

import (
	"strings"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func (i *Informer[T]) Get(namespace, name string) (T, error) {
	inf, ok := i.getInformer(namespace)
	if !ok {
		return i.empty(), ErrNotFound
	}

	cached, ok, err := inf.GetStore().GetByKey(i.key(namespace, name))
	if err != nil {
		return i.empty(), err
	}
	if !ok {
		return i.empty(), ErrNotFound
	}

	resp, err := i.convertEntity(cached)
	if err != nil {
		return i.empty(), err
	}

	return resp, nil
}

func (i *Informer[T]) List(namespaces ...string) ([]T, int) {
	ns := make(map[string]struct{}, len(namespaces))
	for _, n := range namespaces {
		ns[n] = struct{}{}
	}

	unstructuredList, total := i.list(ns)
	list := i.convertList(unstructuredList)

	return list, total
}

func (i *Informer[T]) key(namespace, name string) string {
	if len(namespace) == 0 {
		return name
	}

	return strings.Join([]string{namespace, name}, "/")
}

func (i *Informer[T]) list(ns map[string]struct{}) ([]interface{}, int) {
	var list []interface{}
	total := 0

	i.m.RLock()
	defer i.m.RUnlock()

	for n, inf := range i.informers {
		l := inf.GetStore().List()
		total += len(l)

		_, ok := ns[n]
		if len(ns) == 0 || ok {
			list = append(list, l...)
		}
	}

	return list, total
}

func (i *Informer[T]) convertList(list []interface{}) []T {
	l := make([]T, 0, len(list))

	for _, entity := range list {
		el, err := i.convertEntity(entity)
		if err != nil {
			log.Error().Str("informerType", i.informerType).Msg(err.Error())
			continue
		}

		l = append(l, el)
	}

	return l
}

func (i *Informer[T]) convertEntity(entity interface{}) (T, error) {
	el, ok := entity.(*unstructured.Unstructured)
	if !ok {
		return i.empty(), ErrUnableCast
	}

	t := new(T)
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(el.Object, &t)
	if err != nil || t == nil {
		return i.empty(), ErrUnableToConvert
	}

	return *t, nil
}
