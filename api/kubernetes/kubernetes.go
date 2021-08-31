package kubernetes

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Interface interface {
	// Cluster returns the client for the config, or nil.
	Cluster(name string) (client kubernetes.Interface)
	// Clusters returns an map from config name to clients. You must not modify this.
	Clusters() map[string]kubernetes.Interface
}

type impl map[string]kubernetes.Interface

func (i impl) Cluster(name string) kubernetes.Interface {
	return i[name]
}

func (i impl) Clusters() map[string]kubernetes.Interface {
	return i
}

func NewForConfigs(configs map[string]*rest.Config) (Interface, error) {
	clients := make(impl)
	for contextName, r := range configs {
		i, err := kubernetes.NewForConfig(r)
		if err != nil {
			return clients, err
		}
		clients[contextName] = i
	}
	return clients, nil
}

// NewSingleton creates an instance containing only the in-cluster interface
func NewSingleton(name string, v kubernetes.Interface) Interface {
	return impl{name: v}
}
