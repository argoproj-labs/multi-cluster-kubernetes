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

func status2(w http.ResponseWriter) func(v interface{}, err error) {
	return func(v interface{}, err error) {
		if err != nil {
			status(w)(err)
		} else {
			status(w)(v)
		}
	}
}
func status(w http.ResponseWriter) func(v interface{}) {
	return func(v interface{}) {
		if err, ok := v.(error); ok {

			w.Header().Set("Content-Type", "application/json")
			statusError, ok := err.(*errors.StatusError)
			if ok {
				code := int(statusError.ErrStatus.Code)
				fmt.Printf("%d\n", code)
				w.WriteHeader(code)
				_ = json.NewEncoder(io.MultiWriter(os.Stderr, w)).Encode(statusError.ErrStatus)
			} else {
				fmt.Printf("500\n")
				w.WriteHeader(500)
				_ = json.NewEncoder(io.MultiWriter(os.Stderr, w)).Encode(metav1.Status{
					Status:  metav1.StatusFailure,
					Reason:  errors.ReasonForError(err),
					Message: err.Error(),
				})
			}
		} else {
			fmt.Printf("200\n")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_ = json.NewEncoder(w).Encode(v)
		}
	}
}

func stream(w http.ResponseWriter) func(_watch watch.Interface) {
	return func(_watch watch.Interface) {
		defer _watch.Stop()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Printf("200\n")
		successes := json.NewEncoder(w)
		failures := json.NewEncoder(io.MultiWriter(os.Stderr, w))
		for event := range _watch.ResultChan() {
			v, ok := event.Object.(*unstructured.Unstructured)
			if ok {
				_ = successes.Encode(map[string]interface{}{"type": event.Type, "object": v})
			} else {
				_ = failures.Encode(map[string]interface{}{"type": event.Type, "object": event.Object})
			}
			_, _ = w.Write([]byte("\n"))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}
