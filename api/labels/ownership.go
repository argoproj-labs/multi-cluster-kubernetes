package labels

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func SetOwnership(obj metav1.Object, cluster string, owner metav1.Object, gvk schema.GroupVersionKind) {
	if cluster == "" && obj.GetNamespace() == owner.GetNamespace() {
		obj.SetOwnerReferences([]metav1.OwnerReference{*metav1.NewControllerRef(owner, gvk)})
		return
	}
	v := obj.GetLabels()
	if v == nil {
		v = map[string]string{}
	}
	v[KeyOwnerCluster] = cluster
	v[KeyOwnerNamespace] = owner.GetNamespace()
	v[KeyOwnerName] = owner.GetName()
	obj.SetLabels(v)
}

func GetOwnership(obj metav1.Object) (cluster, namespace, name string, err error) {
	if len(obj.GetOwnerReferences()) > 0 {
		owner := obj.GetOwnerReferences()[0]
		return "", obj.GetNamespace(), owner.Name, nil
	}
	v := obj.GetLabels()
	if _, ok := v[KeyOwnerName]; !ok {
		return "", "", "", fmt.Errorf("ownership information not found in labels")
	}
	return v[KeyOwnerCluster], v[KeyOwnerNamespace], v[KeyOwnerName], nil
}
