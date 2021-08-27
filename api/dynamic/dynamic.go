package dynamic

import (
	"github.com/argoproj-labs/multi-cluster-kubernetes/api/config"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type Interface interface {
	// Config returns the client for the named config, or nil.
	Config(name string) (client dynamic.Interface)
	// InCluster returns the in-cluster client, or nil.
	InCluster() (client dynamic.Interface)
}

type impl map[string]dynamic.Interface

func (i impl) Config(name string) dynamic.Interface {
	return i[name]
}

func (i impl) InCluster() dynamic.Interface {
	return i.Config(config.InCluster)
}

func NewForConfigs(configs map[string]*rest.Config) (Interface, error) {
	clients := make(impl)
	for contextName, r := range configs {
		i, err := dynamic.NewForConfig(r)
		if err != nil {
			return clients, err
		}
		clients[contextName] = i
	}
	return clients, nil
}

// NewInCluster creates an instance containing only the in-cluster interface
func NewInCluster(v dynamic.Interface) Interface {
	return impl{config.InCluster: v}
}
