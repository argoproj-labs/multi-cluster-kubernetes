package dynamic

import (
	"github.com/argoproj-labs/multi-cluster-kubernetes/api/rest"
	"k8s.io/client-go/dynamic"
)

type Interface interface {
	// Cluster returns the client for the cluster, or nil.
	Cluster(clusterName string) (client dynamic.Interface)
	// InCluster returns the in-cluster client, or nil.
	InCluster() (client dynamic.Interface)
}

type impl map[string]dynamic.Interface

func (i impl) Cluster(clusterName string) dynamic.Interface {
	return i[clusterName]
}

func (i impl) InCluster() dynamic.Interface {
	return i.Cluster(rest.InClusterName)
}

func NewForConfigs(configs rest.Configs) (Interface, error) {
	clients := make(impl)
	for clusterName, r := range configs {
		i, err := dynamic.NewForConfig(r)
		if err != nil {
			return clients, err
		}
		clients[clusterName] = i
	}
	return clients, nil
}

// NewInCluster creates an instance containing only the in-cluster interface
func NewInCluster(v dynamic.Interface) Interface {
	return impl{rest.InClusterName: v}
}
