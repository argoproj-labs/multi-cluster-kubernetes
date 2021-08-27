package kubernetes

import (
	"github.com/argoproj-labs/multi-cluster-kubernetes/api/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Interface interface {
	// Config returns the client for the config, or nil.
	Config(name string) (client kubernetes.Interface)
	// InCluster returns the in-cluster client, or nil.
	InCluster() (client kubernetes.Interface)
	// Configs returns an map from config name to clients. You must not modify this.
	Configs() map[string]kubernetes.Interface
}

type impl map[string]kubernetes.Interface

func (i impl) Config(name string) kubernetes.Interface {
	return i[name]
}

func (i impl) InCluster() kubernetes.Interface {
	return i.Config(config.InCluster)
}

func (i impl) Configs() map[string]kubernetes.Interface {
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

// NewInCluster creates an instance containing only the in-cluster interface
func NewInCluster(v kubernetes.Interface) Interface {
	return impl{config.InCluster: v}
}
