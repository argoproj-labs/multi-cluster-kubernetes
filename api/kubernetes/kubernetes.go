package kubernetes

import (
	mcrest "github.com/argoproj-labs/multi-cluster-kubernetes/api/rest"
	"k8s.io/client-go/kubernetes"
)

type Interface interface {
	// Cluster returns the client for the cluster, or nil.
	Cluster(clusterName string) (client kubernetes.Interface)
	// InCluster returns the in-cluster client, or nil.
	InCluster() (client kubernetes.Interface)
	// Clusters returns an map from cluster name to clients. You must not modify this.
	Clusters() map[string]kubernetes.Interface
}

type impl map[string]kubernetes.Interface

func (i impl) Cluster(clusterName string) kubernetes.Interface {
	return i[clusterName]
}

func (i impl) InCluster() kubernetes.Interface {
	return i.Cluster(mcrest.InClusterName)
}

func (i impl) Clusters() map[string]kubernetes.Interface {
	return i
}

func NewForConfigs(configs mcrest.Configs) (Interface, error) {
	clients := make(impl)
	for clusterName, c := range configs {
		i, err := kubernetes.NewForConfig(c)
		if err != nil {
			return clients, err
		}
		clients[clusterName] = i
	}
	return clients, nil
}

// NewInCluster creates an instance containing only the in-cluster interface
func NewInCluster(v kubernetes.Interface) Interface {
	return impl{mcrest.InClusterName: v}
}
