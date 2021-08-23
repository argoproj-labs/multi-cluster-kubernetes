package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj-labs/multi-cluster-kubernetes/api"
	gorillaschema "github.com/gorilla/schema"
	"io"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"net/http"
	"strings"
)

const keyClusterName = "clusterName"

var decoder = gorillaschema.NewDecoder()

func init() {
	decoder.IgnoreUnknownKeys(true)
}

func New(config *rest.Config, namespace string) (func(ctx context.Context) error, error) {
	ctx := context.Background()
	secretInterface := kubernetes.NewForConfigOrDie(config).CoreV1().Secrets(namespace)
	configs, err := api.LoadClusters(ctx, secretInterface)
	if err != nil {
		return nil, err
	}
	clients := make(map[string]dynamic.Interface)
	for clusterName, c := range configs {
		fmt.Printf("%s -> %s\n", clusterName, c.Host)
		clients[clusterName] = dynamic.NewForConfigOrDie(config)
	}
	mux := http.NewServeMux()
	server := server{
		Server:  http.Server{Addr: ":2473", Handler: mux},
		disco:   discovery.NewDiscoveryClientForConfigOrDie(config),
		configs: configs,
		clients: clients,
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			_, _ = io.Copy(io.Discard, r.Body)
			_ = r.Body.Close()
		}()
		fmt.Printf("%s %s ", r.Method, r.URL)
		parts := strings.Split(r.URL.Path, "/")
		switch parts[1] {
		case "api":
			server.api(w, r, parts[2:])
		case "apis":
			server.apis(w, r, parts[2:])
		case "openapi":
			server.openapi(w)
		default:
			status(w)(fmt.Errorf("unknown %q", parts[1]))
		}
	})

	go func() {
		defer runtime.HandleCrash()
		fmt.Printf("starting server on %q\n", server.Addr)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	return func(ctx context.Context) error {
		fmt.Printf("shutting down server\n")
		return server.Shutdown(ctx)
	}, nil
}

type server struct {
	http.Server
	configs map[string]*rest.Config
	clients map[string]dynamic.Interface
	disco   discovery.DiscoveryInterface
}

func (s *server) api(w http.ResponseWriter, r *http.Request, parts []string) {
	switch len(parts) {
	case 0:
		status(w)(metav1.APIVersions{TypeMeta: metav1.TypeMeta{Kind: "APIVersions"}, Versions: []string{}})
	case 1:
		groupVersion := parts[0]
		status2(w)(s.disco.ServerResourcesForGroupVersion(groupVersion))
	default:
		s.apis(w, r, append([]string{""}, parts...))
	}
}

func (s *server) apis(w http.ResponseWriter, r *http.Request, parts []string) {
	switch len(parts) {
	case 0:
		status2(w)(s.disco.ServerGroups())
	case 2:
		status2(w)(s.disco.ServerResourcesForGroupVersion(parts[0] + "/" + parts[1]))
	case 3:
		group := parts[0]
		version := parts[1]
		resource := parts[2]
		gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
		switch r.Method {
		case "DELETE":
			status(w)(s.deleteCollection(r, "", gvr))
		case "GET":
			status2(w)(s.list(r, "", gvr))
		case "POST":
			status2(w)(s.create(r, "", gvr))
		default:
			status(w)(errors.NewMethodNotSupported(gvr.GroupResource(), r.Method))
		}
	case 4:
		group := parts[0]
		version := parts[1]
		resource := parts[2]
		name := parts[3]
		gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
		switch r.Method {
		case "PUT":
			status2(w)(s.update(r, "", gvr))
		case "PATCH":
			status2(w)(s.patch(r, "", name, gvr))
		default:
			status(w)(errors.NewMethodNotSupported(gvr.GroupResource(), r.Method))
		}
	case 5:
		group := parts[0]
		version := parts[1]
		namespace := parts[3]
		resource := parts[4]
		gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
		switch r.Method {
		case "DELETE":
			status(w)(s.deleteCollection(r, namespace, gvr))
		case "GET":
			if r.URL.Query().Get("watch") == "true" {
				clusterName, _watch, err := s.watch(r, namespace, gvr)
				if err != nil {
					status(w)(err)
				} else {
					stream(w)(clusterName, _watch)
				}
			} else {
				status2(w)(s.list(r, namespace, gvr))
			}
		case "POST":
			status2(w)(s.create(r, namespace, gvr))
		default:
			status(w)(errors.NewMethodNotSupported(gvr.GroupResource(), r.Method))
		}
	case 6:
		group := parts[0]
		version := parts[1]
		namespace := parts[3]
		resource := parts[4]
		name := parts[5]
		gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
		switch r.Method {
		case "PUT":
			status2(w)(s.update(r, namespace, gvr))
		case "PATCH":
			status2(w)(s.patch(r, namespace, name, gvr))
		default:
			status(w)(errors.NewMethodNotSupported(gvr.GroupResource(), r.Method))
		}
	default:
		status(w)(errors.NewInternalError(fmt.Errorf("unexpected number of path parts %d", len(parts))))
	}
}

func (s *server) create(r *http.Request, namespace string, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	opts := metav1.CreateOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	obj := &unstructured.Unstructured{}
	if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
		return nil, err
	}
	clusterName := obj.GetLabels()[keyClusterName]
	resourceInterface, err := s.resource(clusterName, gvr, namespace)
	if err != nil {
		return nil, err
	}
	return resourceInterface.Create(r.Context(), obj, opts)
}

func (s *server) client(clusterName string) (dynamic.Interface, error) {
	client, ok := s.clients[clusterName]
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	return client, nil
}

func (s *server) resource(clusterName string, gvr schema.GroupVersionResource, namespace string) (dynamic.ResourceInterface, error) {
	client, err := s.client(clusterName)
	if err != nil {
		return nil, err
	}
	return client.Resource(gvr).Namespace(namespace), nil
}

func (s *server) list(r *http.Request, namespace string, gvr schema.GroupVersionResource) (*unstructured.UnstructuredList, error) {
	opts, clusterName, err := s.listOptions(r)
	if err != nil {
		return nil, err
	}
	resourceInterface, err := s.resource(clusterName, gvr, namespace)
	if err != nil {
		return nil, err
	}
	list, err := resourceInterface.List(r.Context(), opts)
	if err != nil {
		return nil, err
	}
	for _, item := range list.Items {
		item.GetLabels()[keyClusterName] = clusterName
	}
	return list, err
}

func (s *server) watch(r *http.Request, namespace string, gvr schema.GroupVersionResource) (string, watch.Interface, error) {
	opts, clusterName, err := s.listOptions(r)
	if err != nil {
		return "", nil, err
	}
	resourceInterface, err := s.resource(clusterName, gvr, namespace)
	if err != nil {
		return "", nil, err
	}
	w, err := resourceInterface.Watch(r.Context(), opts)
	if err != nil {
		return "", nil, err
	}
	return clusterName, w, nil
}

func (s *server) update(r *http.Request, namespace string, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	opts := metav1.UpdateOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	obj := &unstructured.Unstructured{}
	if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
		return nil, err
	}
	clusterName := obj.GetLabels()[keyClusterName]
	resourceInterface, err := s.resource(clusterName, gvr, namespace)
	if err != nil {
		return nil, err
	}
	delete(obj.GetLabels(), keyClusterName)
	v, err := resourceInterface.Update(r.Context(), obj, opts)
	if err != nil {
		return nil, err
	}
	v.GetLabels()[keyClusterName] = clusterName
	return v, err
}

func (s *server) patch(r *http.Request, namespace, name string, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	opts := metav1.PatchOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	obj := &unstructured.Unstructured{}
	if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
		return nil, err
	}
	clusterName := obj.GetLabels()[keyClusterName]
	resourceInterface, err := s.resource(clusterName, gvr, namespace)
	if err != nil {
		return nil, err
	}
	delete(obj.GetLabels(), keyClusterName)
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	v, err := resourceInterface.Patch(r.Context(), name, types.MergePatchType, data, opts)
	if err != nil {
		return nil, err
	}
	v.GetLabels()[keyClusterName] = clusterName
	return v, err
}

func (s *server) deleteCollection(r *http.Request, namespace string, gvr schema.GroupVersionResource) error {
	opts := metav1.DeleteOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return err
	}
	listOptions, clusterName, err := s.listOptions(r)
	if err != nil {
		return err
	}
	resourceInterface, err := s.resource(clusterName, gvr, namespace)
	if err != nil {
		return err
	}
	return resourceInterface.DeleteCollection(r.Context(), opts, listOptions)
}

func (s *server) listOptions(r *http.Request) (metav1.ListOptions, string, error) {
	listOptions := metav1.ListOptions{}
	if err := decoder.Decode(&listOptions, r.URL.Query()); err != nil {
		return listOptions, "", err
	}
	selector, err := labels.Parse(listOptions.LabelSelector)
	if err != nil {
		return listOptions, "", err
	}
	requirements, _ := selector.Requirements()
	newSelector := labels.NewSelector()
	clusterName := ""
	for _, r := range requirements {
		if r.Key() != keyClusterName {
			newSelector = newSelector.Add(r)
		} else {
			var ok bool
			clusterName, ok = r.Values().PopAny()
			if !ok {
				return listOptions, "", errors.NewBadRequest("invalid clusterName selector")
			}
		}
	}
	listOptions.LabelSelector = newSelector.String()
	return listOptions, clusterName, nil
}

func (s *server) openapi(w http.ResponseWriter) {
	document, err := s.disco.OpenAPISchema()
	if err != nil {
		status(w)(err)
	} else {
		marshal, err := document.XXX_Marshal(nil, true)
		if err != nil {
			status(w)(err)
		}
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/com.github.proto-openapi.spec.v2@v1.0+protobuf")
		_, _ = w.Write(marshal)
	}
}
