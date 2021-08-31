package dynamic

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type Interface interface {
	// Cluster returns the client for the named config, or nil.
	Cluster(name string) (client dynamic.Interface)
}

type impl map[string]dynamic.Interface

func (i impl) Cluster(name string) dynamic.Interface {
	return i[name]
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

// NewSingleton creates an instance containing a single interface
func NewSingleton(name string, v dynamic.Interface) Interface {
	return impl{name: v}
}
