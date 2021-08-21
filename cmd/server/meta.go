package server

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

func setMetaData(v *unstructured.Unstructured, clusterName string) {
	v.SetClusterName(clusterName)
	if v.GetNamespace() == "" && v.GetKind() != "CustomResourceDefinition" {
		v.SetName(join(v.GetClusterName(), v.GetName()))
	} else if v.GetNamespace() != "" {
		v.SetNamespace(join(v.GetClusterName(), v.GetNamespace()))
	}
}
