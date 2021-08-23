package server

import (
	"context"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
	"testing"
)

func TestCustomResourceDefinitions(t *testing.T) {
	defer Setup()()
	restConfig := RestConfig()
	ctx := context.Background()

	resourceInterface := dynamic.NewForConfigOrDie(restConfig).
		Resource(schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"})

	t.Run("DeleteCollection", func(t *testing.T) {
		err := resourceInterface.DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "test"})
		assert.NoError(t, err)
	})
	t.Run("Create", func(t *testing.T) {
		obj := &unstructured.Unstructured{}
		err := yaml.Unmarshal([]byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: tests.argoproj.io
  labels:
    test: "yes"
spec:
  group: argoproj.io
  names:
    kind: Test
    listKind: TestList
    plural: tests
    singular: test
  scope: Namespaced
  versions:
    - name: v1alpha1
      schema:
        openAPIV3Schema:
          properties:
            apiVersion:
              type: string
            kind:
              type: string
            metadata:
              type: object
            spec:
              type: object
              x-kubernetes-map-type: atomic
              x-kubernetes-preserve-unknown-fields: true
          required:
            - metadata
          type: object
      served: true
      storage: true
`), obj)
		assert.NoError(t, err)
		_, err = resourceInterface.Create(ctx, obj, metav1.CreateOptions{})
		assert.NoError(t, err)
	})
	t.Run("Get", func(t *testing.T) {
		_, err := resourceInterface.Get(ctx, "tests.argoproj.io", metav1.GetOptions{})
		assert.NoError(t, err)
	})

	t.Run("Delete", func(t *testing.T) {
		err := resourceInterface.Delete(ctx, "tests.argoproj.io", metav1.DeleteOptions{})
		assert.NoError(t, err)
	})

}
