package dynamic

import (
	"github.com/argoproj-labs/multi-cluster-kubernetes/api/rest"
	"k8s.io/client-go/dynamic"
)

type Interfaces struct {
	value map[string]dynamic.Interface
}

func (i Interfaces) Cluster(clusterName string) (dynamic.Interface, bool) {
	j, ok := i.value[clusterName]
	return j, ok
}

func (i Interfaces) InCluster() dynamic.Interface {
	v, _ := i.Cluster(rest.InClusterName)
	return v
}

func NewForConfigs(configs rest.Configs) (Interfaces, error) {
	clients := Interfaces{value: map[string]dynamic.Interface{}}
	for clusterName, r := range configs {
		i, err := dynamic.NewForConfig(r)
		if err != nil {
			return clients, err
		}
		clients.value[clusterName] = i
	}
	return clients, nil
}
