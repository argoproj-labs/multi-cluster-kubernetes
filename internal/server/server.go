package server

import (
	"context"
	"encoding/json"
	"fmt"
	clusters "github.com/argoproj-labs/multi-cluster-kubernetes-api/internal/clusters"
	gorillaschema "github.com/gorilla/schema"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var decoder = gorillaschema.NewDecoder()

func init() {
	decoder.IgnoreUnknownKeys(true)
}

func New(config *rest.Config, namespace string) (func(ctx context.Context) error, error) {
	secret, err := kubernetes.NewForConfigOrDie(config).CoreV1().Secrets(namespace).Get(context.Background(), "clusters", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	configs := make(map[string]*rest.Config)
	clients := make(map[string]dynamic.Interface)
	for clusterName, data := range secret.Data {
		c := &clusters.Config{}
		if err := json.Unmarshal(data, c); err != nil {
			return nil, err
		}
		fmt.Printf("%s -> %s\n", clusterName, c.Host)
		config := c.RestConfig()
		configs[clusterName] = config
		clients[clusterName] = dynamic.NewForConfigOrDie(config)
	}
	server := server{
		Server:  http.Server{Addr: ":2473"},
		disco:   discovery.NewDiscoveryClientForConfigOrDie(config),
		configs: configs,
		clients: clients,
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
			nok(w, fmt.Errorf("unknown %q", parts[1]))
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
		ok(w, metav1.APIVersions{TypeMeta: metav1.TypeMeta{Kind: "APIVersions"}, Versions: []string{}})
	case 1:
		groupVersion := parts[0]
		list, err := s.disco.ServerResourcesForGroupVersion(groupVersion)
		done(w, list, err)
	default:
		s.apis(w, r, append([]string{""}, parts...))
	}
}

func (s *server) apis(w http.ResponseWriter, r *http.Request, parts []string) {
	switch len(parts) {
	case 0:
		groups, err := s.disco.ServerGroups()
		done(w, groups, err)
	case 2:
		groupVersion, err := s.disco.ServerResourcesForGroupVersion(parts[0] + "/" + parts[1])
		done(w, groupVersion, err)
	case 3:
		group := parts[0]
		version := parts[1]
		resource := parts[2]
		gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
		switch r.Method {
		case "GET":
			list, err := s.clusterList(r, gvr)
			done(w, list, err)
		case "POST":
			list, err := s.clusterCreate(r, gvr)
			done(w, list, err)
		default:
			nok(w, errors.NewMethodNotSupported(gvr.GroupResource(), r.Method))
		}
	case 4:
		group := parts[0]
		version := parts[1]
		resource := parts[2]
		clusterName, name := split(parts[3])
		gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
		switch r.Method {
		case "DELETE":
			err := s.clusterDelete(r, clusterName, name, gvr)
			if err != nil {
				nok(w, err)
			}
		case "GET":
			v, err := s.clusterGet(r, clusterName, name, gvr)
			done(w, v, err)
		case "PUT":
			v, err := s.clusterUpdate(r, clusterName, name, gvr)
			done(w, v, err)
		case "PATCH":
			v, err := s.clusterPatch(r, clusterName, name, gvr)
			done(w, v, err)
		default:
			nok(w, errors.NewMethodNotSupported(gvr.GroupResource(), r.Method))
		}
	case 5:
		group := parts[0]
		version := parts[1]
		clusterName, namespace := split(parts[3])
		resource := parts[4]
		gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
		switch r.Method {
		case "DELETE":
			err := s.deleteCollection(r, clusterName, namespace, gvr)
			if err != nil {
				nok(w, err)
			}
		case "GET":
			if r.URL.Query().Get("watch") == "true" {
				_watch, err := s.watch(r, clusterName, namespace, gvr)
				if err != nil {
					nok(w, err)
				} else {
					stream(w, _watch, clusterName)
				}
			} else {
				list, err := s.list(r, clusterName, namespace, gvr)
				done(w, list, err)
			}
		case "POST":
			v, err := s.create(r, clusterName, namespace, gvr)
			done(w, v, err)
		default:
			nok(w, errors.NewMethodNotSupported(gvr.GroupResource(), r.Method))
		}
	case 6:
		group := parts[0]
		version := parts[1]
		clusterName, namespace := split(parts[3])
		resource := parts[4]
		name := parts[5]
		gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
		switch r.Method {
		case "GET":
			v, err := s.get(r, clusterName, namespace, name, gvr)
			done(w, v, err)
		case "PUT":
			v, err := s.update(r, clusterName, namespace, name, gvr)
			done(w, v, err)
		case "PATCH":
			v, err := s.patch(r, clusterName, namespace, name, gvr)
			done(w, v, err)
		case "DELETE":
			err := s.delete(r, clusterName, namespace, name, gvr)
			if err != nil {
				nok(w, err)
			}
		default:
			nok(w, errors.NewMethodNotSupported(gvr.GroupResource(), r.Method))
		}
	default:
		nok(w, errors.NewInternalError(fmt.Errorf("unexpected number of path parts %d", len(parts))))
	}
}

func (s *server) create(r *http.Request, clusterName, namespace string, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	opts := metav1.CreateOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	obj := &unstructured.Unstructured{}
	if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
		return nil, err
	}
	obj.SetNamespace(namespace)
	switch gvr.Resource {
	case "events":
		if err := unstructured.SetNestedField(obj.Object, namespace, "involvedObject", "namespace"); err != nil {
			return nil, err
		}
	}
	client, ok := s.clients[clusterName]
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	v, err := client.Resource(gvr).Namespace(namespace).Create(r.Context(), obj, opts)
	if err != nil {
		return nil, err
	}
	setMetaData(v, clusterName)
	return v, nil
}

func (s *server) clusterList(r *http.Request, gvr schema.GroupVersionResource) (*unstructured.UnstructuredList, error) {
	var clusterList *unstructured.UnstructuredList
	for clusterName := range s.clients {
		list, err := s.list(r, clusterName, "", gvr)
		if err != nil {
			return nil, err
		}
		if clusterList == nil {
			clusterList = list
		} else {
			clusterList.Items = append(clusterList.Items, list.Items...)
		}
	}
	return clusterList, nil
}

func (s *server) clusterDelete(r *http.Request, clusterName, name string, gvr schema.GroupVersionResource) error {
	opts := metav1.DeleteOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return err
	}
	client, ok := s.clients[clusterName]
	if !ok {
		return errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	return client.Resource(gvr).Delete(r.Context(), name, opts)
}

func (s *server) clusterGet(r *http.Request, clusterName, name string, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	opts := metav1.GetOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	client, ok := s.clients[clusterName]
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	v, err := client.Resource(gvr).Get(r.Context(), name, opts)
	if err != nil {
		return nil, err
	}
	setMetaData(v, clusterName)
	return v, nil
}

func (s *server) clusterCreate(r *http.Request, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	opts := metav1.CreateOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	obj := &unstructured.Unstructured{}
	if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
		return nil, err
	}
	clusterName, name := split(obj.GetName())
	obj.SetName(name)
	client, ok := s.clients[clusterName]
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	v, err := client.Resource(gvr).Create(r.Context(), obj, opts)
	if err != nil {
		return nil, err
	}
	setMetaData(v, clusterName)
	return v, nil
}

func (s *server) clusterUpdate(r *http.Request, clusterName, name string, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	opts := metav1.UpdateOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	obj := &unstructured.Unstructured{}
	if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
		return nil, err
	}
	obj.SetName(name)
	client, ok := s.clients[clusterName]
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	v, err := client.Resource(gvr).Update(r.Context(), obj, opts)
	if err != nil {
		return nil, err
	}
	setMetaData(v, clusterName)
	return v, nil
}

func (s *server) clusterPatch(r *http.Request, clusterName, name string, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	opts := metav1.PatchOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	client, ok := s.clients[clusterName]
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	v, err := client.Resource(gvr).Patch(r.Context(), name, types.MergePatchType, data, opts)
	if err != nil {
		return nil, err
	}
	setMetaData(v, clusterName)
	return v, nil
}

func (s *server) list(r *http.Request, clusterName, namespace string, gvr schema.GroupVersionResource) (*unstructured.UnstructuredList, error) {
	opts := metav1.ListOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	client, ok := s.clients[clusterName]
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	list, err := client.Resource(gvr).Namespace(namespace).List(r.Context(), opts)
	if err != nil {
		return nil, err
	}
	for _, v := range list.Items {
		setMetaData(&v, clusterName)
	}
	return list, nil
}

func (s *server) watch(r *http.Request, clusterName, namespace string, gvr schema.GroupVersionResource) (watch.Interface, error) {
	opts := metav1.ListOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	client, ok := s.clients[clusterName]
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	return client.Resource(gvr).Namespace(namespace).Watch(r.Context(), opts)
}

func (s *server) get(r *http.Request, clusterName, namespace, name string, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	opts := metav1.GetOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	client, ok := s.clients[clusterName]
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	v, err := client.Resource(gvr).Namespace(namespace).Get(r.Context(), name, opts)
	if err != nil {
		return nil, err
	}
	setMetaData(v, clusterName)
	return v, nil
}

func (s *server) update(r *http.Request, clusterName, namespace, name string, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	opts := metav1.UpdateOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	obj := &unstructured.Unstructured{}
	if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
		return nil, err
	}
	obj.SetNamespace(namespace)
	obj.SetName(name)
	client, ok := s.clients[clusterName]
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	v, err := client.Resource(gvr).Namespace(namespace).Update(r.Context(), obj, opts)
	if err != nil {
		return nil, err
	}
	setMetaData(v, clusterName)
	return v, nil
}

func (s *server) patch(r *http.Request, clusterName, namespace, name string, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	opts := metav1.PatchOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	client, ok := s.clients[clusterName]
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	v, err := client.Resource(gvr).Namespace(namespace).Patch(r.Context(), name, types.MergePatchType, data, opts)
	if err != nil {
		return nil, err
	}
	setMetaData(v, clusterName)
	return v, nil
}

func (s *server) delete(r *http.Request, clusterName, namespace, name string, gvr schema.GroupVersionResource) error {
	opts := metav1.DeleteOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return err
	}
	client, ok := s.clients[clusterName]
	if !ok {
		return errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	return client.Resource(gvr).Namespace(namespace).Delete(r.Context(), name, opts)
}

func (s *server) deleteCollection(r *http.Request, clusterName, namespace string, gvr schema.GroupVersionResource) error {
	opts := metav1.DeleteOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return err
	}
	listOptions := metav1.ListOptions{}
	if err := decoder.Decode(&listOptions, r.URL.Query()); err != nil {
		return err
	}
	client, ok := s.clients[clusterName]
	if !ok {
		return errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	return client.Resource(gvr).Namespace(namespace).DeleteCollection(r.Context(), opts, listOptions)
}

func (s *server) openapi(w http.ResponseWriter) {
	document, err := s.disco.OpenAPISchema()
	if err != nil {
		nok(w, err)
	} else {
		marshal, err := document.XXX_Marshal(nil, true)
		if err != nil {
			nok(w, err)
		}
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/com.github.proto-openapi.spec.v2@v1.0+protobuf")
		_, _ = w.Write(marshal)
	}
}

func (s *server) createSubResource(w http.ResponseWriter, r *http.Request, clusterName, namespace, name, subresource string, gvr schema.GroupVersionResource) error {
	restConfig, ok := s.configs[clusterName]
	if !ok {
		return errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	transport, upgrader, err := spdy.RoundTripperFor(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create round-tripper: %w", err)
	}
	endpoint := fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/portforward", restConfig.Host, namespace, name)
	fmt.Printf("%s\n", endpoint)
	x, err := url.Parse(endpoint)
	if err != nil {
		return err
	}
	w.Header().Set("Upgrade", "SPDY/3.1")
	w.Header().Set("Connection", "Upgrade")
	w.WriteHeader(http.StatusSwitchingProtocols)
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", x)
	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	forwarder, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", 9090, 9090)}, stopChan, readyChan, w, w)
	if err != nil {
		return fmt.Errorf("failed to create new port-forward: %w", err)
	}
	go func() {
		defer runtime.HandleCrash()
		if err := forwarder.ForwardPorts(); err != nil {
			log.Fatal()
		}
	}()
	<-readyChan
	<-r.Context().Done()
	stopChan <- struct{}{}
	forwarder.Close()
	return r.Context().Err()
}
