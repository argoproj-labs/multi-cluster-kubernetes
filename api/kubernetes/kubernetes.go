package kubernetes

import (
	"github.com/argoproj-labs/multi-cluster-kubernetes/api/config"
	"k8s.io/client-go/kubernetes"
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
	return i.Config(config.InClusterName)
}

func (i impl) Configs() map[string]kubernetes.Interface {
	return i
}

func NewForConfigs(configs config.Configs) (Interface, error) {
	clients := make(impl)
	for clusterName, c := range configs {
		restConfig, err := c.ClientConfig()
		if err != nil {
			return clients, err
		}
		i, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			return clients, err
		}
		clients[clusterName] = i
	}
	return clients, nil
}

// NewInCluster creates an instance containing only the in-cluster interface
func NewInCluster(v kubernetes.Interface) Interface {
	return impl{config.InClusterName: v}
}
