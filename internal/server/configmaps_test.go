package server

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"testing"
)

func TestResource(t *testing.T) {
	defer Setup()()
	restConfig := RestConfig()
	ctx := context.Background()

	resourceInterface := dynamic.NewForConfigOrDie(restConfig).
		Resource(schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}).
		Namespace("default.default")

	t.Run("DeleteCollection", func(t *testing.T) {
		err := resourceInterface.DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "test"})
		assert.NoError(t, err)
	})
	t.Run("Create", func(t *testing.T) {
		obj := &unstructured.Unstructured{}
		obj.SetKind("ConfigMap")
		obj.SetAPIVersion("v1")
		obj.SetName("test")
		obj.SetLabels(map[string]string{"test": "yes"})
		_, err := resourceInterface.Create(ctx, obj, metav1.CreateOptions{})
		assert.NoError(t, err)
	})
	t.Run("List", func(t *testing.T) {
		list, err := resourceInterface.List(ctx, metav1.ListOptions{LabelSelector: "test"})
		if assert.NoError(t, err) {
			assert.Len(t, list.Items, 1)
		}
	})
	var resourceVersion string
	t.Run("Get", func(t *testing.T) {
		item, err := resourceInterface.Get(ctx, "test", metav1.GetOptions{})
		if assert.NoError(t, err) {
			assert.Equal(t, "test", item.GetName())
			assert.Equal(t, "default.default", item.GetNamespace())
			resourceVersion = item.GetResourceVersion()
		}
	})
	t.Run("Update", func(t *testing.T) {
		obj := &unstructured.Unstructured{}
		obj.SetKind("ConfigMap")
		obj.SetAPIVersion("v1")
		obj.SetName("test")
		obj.SetResourceVersion(resourceVersion)
		obj.SetAnnotations(map[string]string{"updated": "yes"})
		updated, err := resourceInterface.Update(ctx, obj, metav1.UpdateOptions{})
		if assert.NoError(t, err) {
			assert.Len(t, updated.GetAnnotations(), 1)
		}
	})
	t.Run("Patch", func(t *testing.T) {
		obj := &unstructured.Unstructured{}
		obj.SetKind("ConfigMap")
		obj.SetAPIVersion("v1")
		obj.SetName("test")
		obj.SetAnnotations(map[string]string{"patched": "yes"})
		data, _ := json.Marshal(obj)
		updated, err := resourceInterface.Patch(ctx, "test", types.MergePatchType, data, metav1.PatchOptions{})
		if assert.NoError(t, err) {
			assert.Len(t, updated.GetAnnotations(), 2)
		}
	})
	t.Run("Delete", func(t *testing.T) {
		err := resourceInterface.Delete(ctx, "test", metav1.DeleteOptions{})
		assert.NoError(t, err)
	})
}
