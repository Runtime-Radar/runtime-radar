package inventory

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/discovery"
)

type resourceClient struct{}

type Resourcer interface {
	Get(d discovery.DiscoveryInterface) ([]schema.GroupVersionResource, error)
}

func (r *resourceClient) Get(d discovery.DiscoveryInterface) ([]schema.GroupVersionResource, error) {
	serverResources, err := d.ServerPreferredResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return nil, fmt.Errorf("can't get server resources: %w", err)
	}
	if err != nil {
		log.Warn().Err(err).Msg("some API groups failed discovery, continuing with partial list")
	}

	var gvrs []schema.GroupVersionResource
	for _, list := range serverResources {
		if len(list.APIResources) == 0 {
			continue
		}

		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			log.Error().Err(err).Msgf("error parsing group version: '%s'", list.GroupVersion)
			continue
		}

		for _, resource := range list.APIResources {
			if len(resource.Verbs) == 0 {
				continue
			}

			if !sets.NewString(resource.Verbs...).HasAll("watch", "list") {
				continue
			}

			gvrs = append(gvrs, gv.WithResource(resource.Name))
		}
	}

	return gvrs, nil
}
