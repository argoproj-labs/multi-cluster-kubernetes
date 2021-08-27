package cache

import (
	"fmt"
	"github.com/argoproj-labs/multi-cluster-kubernetes/api/labels"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func SplitMetaNamespaceKey(key string) (cluster, namespace, name string, err error) {
	parts := strings.Split(key, "/")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("expected %q to have 3 parts", key)
	}
	return parts[0], parts[1], parts[2], err
}

func JoinMetaNamespaceKey(cluster, namespace, name string) string {
	return cluster + "/" + namespace + "/" + name
}

func MetaNamespaceKeyFunc(obj interface{}) (string, error) {
	m, err := meta.Accessor(obj)
	if err != nil {
		return "", fmt.Errorf("object has no meta: %w", err)
	}
	return cluster(m) + "/" + m.GetNamespace() + "/" + m.GetName(), nil
}

func cluster(meta metav1.Object) string {
	if annotations := meta.GetAnnotations(); annotations != nil {
		return annotations[labels.KeyCluster]
	}
	return ""
}
