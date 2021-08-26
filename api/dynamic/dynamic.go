package dynamic

import (
	"github.com/argoproj-labs/multi-cluster-kubernetes/api/config"
	"k8s.io/client-go/dynamic"
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
	return i.Config(config.InClusterName)
}

func NewForConfigs(configs config.Configs) (Interface, error) {
	clients := make(impl)
	for configName, r := range configs {
		restConfig, err := r.ClientConfig()
		if err != nil {
			return clients, err
		}
		i, err := dynamic.NewForConfig(restConfig)
		if err != nil {
			return clients, err
		}
		clients[configName] = i
	}
	return clients, nil
}

// NewInCluster creates an instance containing only the in-cluster interface
func NewInCluster(v dynamic.Interface) Interface {
	return impl{config.InClusterName: v}
}
