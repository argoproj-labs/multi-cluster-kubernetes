package labels

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func TestSetOwnership(t *testing.T) {
	t.Run("SameClusterAndNamespace", func(t *testing.T) {
		obj := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Namespace: "ns"},
		}
		SetOwnership(obj, "", &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "n", Namespace: "ns"},
		}, schema.GroupVersionKind{})
		assert.Len(t, obj.OwnerReferences, 1)
	})
	t.Run("DifferentNamespace", func(t *testing.T) {
		obj := &corev1.Pod{}
		SetOwnership(obj, "", &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "n", Namespace: "ns"},
		}, schema.GroupVersionKind{})
		assert.Len(t, obj.OwnerReferences, 0)
		assert.Equal(t, map[string]string{
			KeyOwnerCluster:   "",
			KeyOwnerNamespace: "ns",
			KeyOwnerName:      "n",
		}, obj.GetLabels())
	})
	t.Run("DifferentCluster", func(t *testing.T) {
		obj := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "ns"}}
		SetOwnership(obj, "cn", &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "n", Namespace: "ns"},
		}, schema.GroupVersionKind{})
		assert.Len(t, obj.OwnerReferences, 0)
		assert.Equal(t, map[string]string{
			KeyOwnerCluster:   "cn",
			KeyOwnerNamespace: "ns",
			KeyOwnerName:      "n",
		}, obj.GetLabels())
	})
}

func TestGetOwnership(t *testing.T) {
	t.Run("Orphan", func(t *testing.T) {
		_, _, _, err := GetOwnership(&corev1.Pod{})
		assert.Error(t, err)
	})
	t.Run("SameClusterAndNamespace", func(t *testing.T) {
		cluster, namespace, name, err := GetOwnership(&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:       "ns",
				OwnerReferences: []metav1.OwnerReference{{Name: "n"}},
			},
		})
		assert.NoError(t, err)
		assert.Empty(t, cluster)
		assert.Equal(t, "ns", namespace)
		assert.Equal(t, "n", name)
	})
	t.Run("DifferentNamespace", func(t *testing.T) {
		cluster, namespace, name, err := GetOwnership(&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					KeyOwnerCluster:   "",
					KeyOwnerNamespace: "ns",
					KeyOwnerName:      "n",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, "", cluster)
		assert.Equal(t, "ns", namespace)
		assert.Equal(t, "n", name)
	})
	t.Run("DifferentCluster", func(t *testing.T) {
		cluster, namespace, name, err := GetOwnership(&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					KeyOwnerCluster:   "cn",
					KeyOwnerNamespace: "",
					KeyOwnerName:      "n",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, "cn", cluster)
		assert.Equal(t, "", namespace)
		assert.Equal(t, "n", name)
	})
}
