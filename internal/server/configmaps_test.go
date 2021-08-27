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

	const cluster = "k3d-k3s-default"
	resourceInterface := dynamic.NewForConfigOrDie(restConfig).
		Resource(schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}).
		Namespace("default")

	t.Run("DeleteCollection", func(t *testing.T) {
		err := resourceInterface.DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "cluster="+ cluster +",test"})
		assert.NoError(t, err)
	})
	var resourceVersion string
	t.Run("Create", func(t *testing.T) {
		obj := &unstructured.Unstructured{}
		obj.SetKind("ConfigMap")
		obj.SetAPIVersion("v1")
		obj.SetName("test")
		obj.SetLabels(map[string]string{"test": "yes", "cluster": cluster})
		created, err := resourceInterface.Create(ctx, obj, metav1.CreateOptions{})
		if assert.NoError(t, err) {
			resourceVersion = created.GetResourceVersion()
		}
	})
	t.Run("List", func(t *testing.T) {
		list, err := resourceInterface.List(ctx, metav1.ListOptions{LabelSelector: "cluster="+ cluster +",test"})
		if assert.NoError(t, err) {
			assert.Len(t, list.Items, 1)
		}
	})
	t.Run("Update", func(t *testing.T) {
		obj := &unstructured.Unstructured{}
		obj.SetKind("ConfigMap")
		obj.SetAPIVersion("v1")
		obj.SetName("test")
		obj.SetResourceVersion(resourceVersion)
		obj.SetAnnotations(map[string]string{"updated": "yes"})
		obj.SetLabels(map[string]string{"cluster": cluster, "test": "yes"})
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
		obj.SetLabels(map[string]string{"cluster": cluster})
		data, _ := json.Marshal(obj)
		updated, err := resourceInterface.Patch(ctx, "test", types.MergePatchType, data, metav1.PatchOptions{})
		if assert.NoError(t, err) {
			assert.Len(t, updated.GetAnnotations(), 2)
		}
	})
}
