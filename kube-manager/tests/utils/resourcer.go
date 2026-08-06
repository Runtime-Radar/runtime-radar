package utils

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

type FakeResourcer struct {
	Resources []schema.GroupVersionResource
}

func (fr *FakeResourcer) Get(_ discovery.DiscoveryInterface) ([]schema.GroupVersionResource, error) {
	return fr.Resources, nil
}
