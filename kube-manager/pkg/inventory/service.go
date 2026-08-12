package inventory

import (
	"fmt"
	"time"

	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/informers"
	"github.com/runtime-radar/runtime-radar/kube-manager/pkg/updater"
	"k8s.io/client-go/dynamic"
	k8s_informers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type Service interface {
	Run(stop chan struct{}) error
	Shutdown()
}

type Inventory struct {
	nsFactory  k8s_informers.SharedInformerFactory
	nsInformer cache.SharedIndexInformer

	staticClient  kubernetes.Interface
	dynamicClient dynamic.Interface

	dynamicFactories FactoryBuffer
	resourcer        Resourcer
	updater          updater.ConfigUpdater

	syncInterval time.Duration
	informers    []informers.Setter
}

func New(
	updater updater.ConfigUpdater,
	config *rest.Config,
	syncDuration time.Duration,
	infs ...informers.Setter,
) (Service, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("can't create k8s client: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("can't create dynamic k8s client: %w", err)
	}

	return NewWithClients(updater, clientset, dynamicClient, &resourceClient{}, syncDuration, infs...)
}

func NewWithClients(
	updater updater.ConfigUpdater,
	staticClient kubernetes.Interface,
	dynamicClient dynamic.Interface,
	resourceClient Resourcer,
	syncDuration time.Duration,
	infs ...informers.Setter,
) (*Inventory, error) {
	f := k8s_informers.NewSharedInformerFactory(staticClient, syncDuration)

	i := &Inventory{
		informers:        infs,
		updater:          updater,
		nsFactory:        f,
		staticClient:     staticClient,
		dynamicClient:    dynamicClient,
		resourcer:        resourceClient,
		dynamicFactories: NewFactoryBuffer(),
		syncInterval:     syncDuration,
	}
	var err error
	i.nsInformer, err = i.addNamespaceInformer()
	if err != nil {
		return nil, fmt.Errorf("can't add namespace informer: %w", err)
	}

	if err := i.addFactory(defaultNamespace); err != nil {
		return nil, fmt.Errorf("can't add factory for non-namespace entities: %w", err)
	}

	updater.SetOnUpdateFunc(func() error {
		return i.onUpdate()
	})

	return i, nil
}

func (i *Inventory) Run(stopCh chan struct{}) error {
	i.nsFactory.Start(stopCh)
	i.nsFactory.WaitForCacheSync(stopCh)

	if err := i.onUpdate(); err != nil {
		return err
	}

	return nil
}

func (i *Inventory) Shutdown() {
	i.dynamicFactories.removeAll()
	i.nsFactory.Shutdown()
}

func (i *Inventory) onUpdate() error {
	allow, deny := i.updater.Maps()
	namespaces, _ := i.List(nil)

	if len(allow) > 0 {
		for _, ns := range namespaces {
			_, started := i.dynamicFactories.get(ns.Name)
			_, isAllow := allow[ns.Name]

			if !started && isAllow {
				if err := i.addFactory(ns.Name); err != nil {
					return fmt.Errorf("can't add factory: %w", err)
				}

				continue
			}

			if started && !isAllow {
				if err := i.removeFactory(ns.Name); err != nil {
					return fmt.Errorf("can't remove factory: %w", err)
				}
			}
		}

		return nil
	}

	for _, ns := range namespaces {
		_, isDeny := deny[ns.Name]
		_, started := i.dynamicFactories.get(ns.Name)

		if !started && !isDeny {
			if err := i.addFactory(ns.Name); err != nil {
				return fmt.Errorf("can't add factory: %w", err)
			}

			continue
		}

		if started && isDeny {
			if err := i.removeFactory(ns.Name); err != nil {
				return fmt.Errorf("can't remove factory: %w", err)
			}
		}
	}

	return nil
}
