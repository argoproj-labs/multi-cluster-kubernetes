package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/argoproj-labs/multi-cluster-kubernetes/api/labels"
)

func TestSplitMetaNamespaceKey(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		_, _, _, err := SplitMetaNamespaceKey("")
		assert.Error(t, err)
	})
	t.Run("Empty", func(t *testing.T) {
		cluster, namespace, name, err := SplitMetaNamespaceKey("a/b/c")
		assert.NoError(t, err)
		assert.Equal(t, "a", cluster)
		assert.Equal(t, "b", namespace)
		assert.Equal(t, "c", name)
	})
}

func TestJoinMetaNamespaceKey(t *testing.T) {
	key := JoinMetaNamespaceKey("a", "b", "c")
	assert.Equal(t, "a/b/c", key)
}

func TestMetaNamespaceKeyFunc(t *testing.T) {
	t.Run("NoCluster", func(t *testing.T) {
		key, err := MetaNamespaceKeyFunc(&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns",
				Name:      "n",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, "/ns/n", key)
	})
	t.Run("Cluster", func(t *testing.T) {
		key, err := MetaNamespaceKeyFunc(&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "ns",
				Name:      "n",
				Annotations: map[string]string{
					labels.KeyCluster: "cn",
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, "cn/ns/n", key)
	})
}
