package server

import (
	"encoding/json"
	"fmt"
	"io"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"net/http"
	"os"
)

func done(w http.ResponseWriter, v interface{}, err error) {
	if err != nil {
		serverError(w, err)
	} else {
		ok(w, v)
	}
}

func ok(w http.ResponseWriter, v interface{}) {
	fmt.Printf("200\n")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	_ = json.NewEncoder(io.MultiWriter(os.Stdout, w)).Encode(v)
}

func stream(w http.ResponseWriter, watch watch.Interface, clusterName, resource string) {
	defer watch.Stop()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	encoder := json.NewEncoder(io.MultiWriter(os.Stdout, w))
	for event := range watch.ResultChan() {
		v, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			_ = encoder.Encode(errors.FromObject(event.Object))
			return
		}
		setMetaData(v, clusterName, resource)
		_ = encoder.Encode(v)
	}
}

func serverError(w http.ResponseWriter, err error) {
	fmt.Printf("500\n")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)
	_ = json.NewEncoder(io.MultiWriter(os.Stderr, w)).Encode(metav1.Status{
		Status:  "Failure",
		Reason:  errors.ReasonForError(err),
		Message: err.Error(),
	})
}
