package server

import (
	"context"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"testing"
)

func TestNamespaces(t *testing.T) {
	defer Setup()()
	restConfig := RestConfig()
	ctx := context.Background()
	resourceInterface := dynamic.NewForConfigOrDie(restConfig).
		Resource(schema.GroupVersionResource{Version: "v1", Resource: "namespaces"})
	t.Run("Create", func(t *testing.T) {
		obj := &unstructured.Unstructured{}
		obj.SetKind("Namespace")
		obj.SetAPIVersion("v1")
		obj.SetName("default.test")
		created, err := resourceInterface.Create(ctx, obj, metav1.CreateOptions{})
		if assert.NoError(t, err) {
			assert.Equal(t, "default.test", created.GetName())
		}
	})
	t.Run("Get", func(t *testing.T) {
		item, err := resourceInterface.Get(ctx, "test", metav1.GetOptions{})
		if assert.NoError(t, err) {
			assert.Equal(t, "default.test", item.GetName())
		}
	})
	t.Run("Delete", func(t *testing.T) {
		err := resourceInterface.Delete(ctx, "default.test", metav1.DeleteOptions{})
		assert.NoError(t, err)
	})
}
