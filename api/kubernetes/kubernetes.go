package kubernetes

import (
	restapi "github.com/argoproj-labs/multi-cluster-kubernetes/api/rest"
	"k8s.io/client-go/kubernetes"
)

type Interfaces struct {
	value map[string]kubernetes.Interface
}

func (i Interfaces) Cluster(clusterName string) (kubernetes.Interface, bool) {
	j, ok := i.value[clusterName]
	return j, ok
}

func (i Interfaces) InCluster() kubernetes.Interface {
	v, _ := i.Cluster(restapi.InClusterName)
	return v
}

func (i Interfaces) Clusters() map[string]kubernetes.Interface {
	return i.value
}

func NewForConfigs(configs restapi.Configs) (Interfaces, error) {
	clients := Interfaces{value: map[string]kubernetes.Interface{}}
	for clusterName, c := range configs {
		i, err := kubernetes.NewForConfig(c)
		if err != nil {
			return clients, err
		}
		clients.value[clusterName] = i
	}
	return clients, nil
}
