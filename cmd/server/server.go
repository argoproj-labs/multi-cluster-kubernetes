package server

import (
	"encoding/json"
	"fmt"
	gorillaschema "github.com/gorilla/schema"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"net/http"
)

var decoder = gorillaschema.NewDecoder()

type server struct {
	http.Server
	clients map[string]dynamic.Interface
	disco   *discovery.DiscoveryClient
}

func (s *server) api(w http.ResponseWriter, r *http.Request, parts []string, disco *discovery.DiscoveryClient) {
	switch len(parts) {
	case 2:
		ok(w, metav1.APIVersions{TypeMeta: metav1.TypeMeta{Kind: "APIVersions"}, Versions: []string{}})
	case 3:
		list, err := disco.ServerResourcesForGroupVersion(parts[2])
		done(w, list, err)
	case 4:
		list, err := s.clusterList(r, "", parts[2], parts[3])
		done(w, list, err)
	case 5:
		clusterName, name := split(parts[4])
		v, err := s.clusterGet(r, "", parts[2], clusterName, parts[3], name)
		done(w, v, err)
	case 6:
		clusterName, namespace := split(parts[4])
		resource := parts[5]
		if r.URL.Query().Get("watch") == "true" {
			watch, err := s.watch(r, "", parts[2], clusterName, namespace, resource)
			if err != nil {
				serverError(w, err)
			} else {
				stream(w, watch, clusterName, resource)
			}
		} else {
			list, err := s.list(r, "", parts[2], clusterName, namespace, resource)
			done(w, list, err)
		}
	case 7:
		clusterName, namespace := split(parts[4])
		v, err := s.get(r, "", parts[2], clusterName, namespace, parts[5], parts[6])
		done(w, v, err)
	default:
		serverError(w, errors.NewInternalError(fmt.Errorf("unexpected number of path parts %d", len(parts))))
	}
}

func (s *server) apis(w http.ResponseWriter, r *http.Request, parts []string, disco *discovery.DiscoveryClient) {
	switch len(parts) {
	case 2:
		groups, err := disco.ServerGroups()
		done(w, groups, err)
	case 4:
		groupVersion, err := disco.ServerResourcesForGroupVersion(parts[2] + "/" + parts[3])
		done(w, groupVersion, err)
	case 5:
		list, err := s.clusterList(r, parts[2], parts[3], parts[4])
		done(w, list, err)
	case 6:
		clusterName, name := split(parts[4])
		v, err := s.clusterGet(r, "", parts[2], clusterName, parts[3], name)
		done(w, v, err)
	case 7:
		switch r.Method {
		case "GET":
			clusterName, namespace := split(parts[5])
			list, err := s.list(r, parts[2], parts[3], clusterName, namespace, parts[6])
			done(w, list, err)
		case "POST":
			clusterName, namespace := split(parts[5])
			v, err := s.create(r, parts[2], parts[3], clusterName, namespace, parts[6])
			done(w, v, err)
		default:
			serverError(w, errors.NewBadRequest("unexpected method"))
		}
	case 8:
		clusterName, namespace := split(parts[5])
		v, err := s.get(r, parts[2], parts[3], clusterName, namespace, parts[6], parts[7])
		done(w, v, err)
	default:
		serverError(w, errors.NewInternalError(fmt.Errorf("unexpected number of path parts %d", len(parts))))
	}
}

func (s *server) create(r *http.Request, group, version, clusterName, namespace, resource string) (*unstructured.Unstructured, error) {
	fmt.Printf("create %s/%s/%s.%s/%s\n", group, version, clusterName, namespace, resource)
	opts := metav1.CreateOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	obj := &unstructured.Unstructured{}
	if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
		return nil, err
	}
	obj.SetNamespace(namespace)
	client, ok := s.clients[clusterName]
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	v, err := client.Resource(schema.GroupVersionResource{Group: group, Version: version, Resource: resource}).Namespace(namespace).Create(r.Context(), obj, opts)
	if err != nil {
		return nil, err
	}
	setMetaData(v, clusterName, resource)
	return v, nil
}

func (s *server) clusterList(r *http.Request, group, version, resource string) (*unstructured.UnstructuredList, error) {
	fmt.Printf("cluster list %s/%s/%s\n", group, version, resource)
	var clusterList *unstructured.UnstructuredList
	for clusterName := range s.clients {
		list, err := s.list(r, group, version, clusterName, "", resource)
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

func (s *server) list(r *http.Request, group, version, clusterName, namespace, resource string) (*unstructured.UnstructuredList, error) {
	fmt.Printf("list %s/%s/%s.%s/%s\n", group, version, clusterName, namespace, resource)
	opts := metav1.ListOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	client, ok := s.clients[clusterName]
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	list, err := client.Resource(schema.GroupVersionResource{Group: group, Version: version, Resource: resource}).Namespace(namespace).List(r.Context(), opts)
	if err != nil {
		return nil, err
	}
	for _, v := range list.Items {
		setMetaData(&v, clusterName, resource)
	}
	return list, nil
}

func (s *server) watch(r *http.Request, group, version, clusterName, namespace, resource string) (watch.Interface, error) {
	fmt.Printf("watch %s/%s/%s.%s/%s\n", group, version, clusterName, namespace, resource)
	opts := metav1.ListOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	client, ok := s.clients[clusterName]
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	return client.Resource(schema.GroupVersionResource{Group: group, Version: version, Resource: resource}).Namespace(namespace).Watch(r.Context(), opts)
}

func setMetaData(v *unstructured.Unstructured, clusterName string, resource string) {
	v.SetClusterName(clusterName)
	if resource == "namespaces" {
		v.SetName(join(v.GetClusterName(), v.GetName()))
	} else {
		v.SetNamespace(join(v.GetClusterName(), v.GetNamespace()))
	}
}

func (s *server) get(r *http.Request, group, version, clusterName, namespace, resource, name string) (*unstructured.Unstructured, error) {
	fmt.Printf("get %s/%s/%s.%s/%s\n", group, version, clusterName, namespace, resource)
	opts := metav1.GetOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	client, ok := s.clients[clusterName]
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	v, err := client.Resource(schema.GroupVersionResource{Group: group, Version: version, Resource: resource}).Namespace(namespace).Get(r.Context(), name, opts)
	if err != nil {
		return nil, err
	}
	setMetaData(v, clusterName, resource)
	return v, nil
}

func (s *server) clusterGet(r *http.Request, group, version, clusterName, resource, name string) (*unstructured.Unstructured, error) {
	fmt.Printf("get %s/%s/%s/%s\n", group, version, clusterName, resource)
	opts := metav1.GetOptions{}
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		return nil, err
	}
	client, ok := s.clients[clusterName]
	if !ok {
		return nil, errors.NewBadRequest(fmt.Sprintf("unknown cluster %q", clusterName))
	}
	v, err := client.Resource(schema.GroupVersionResource{Group: group, Version: version, Resource: resource}).Get(r.Context(), name, opts)
	if err != nil {
		return nil, err
	}
	setMetaData(v, clusterName, resource)
	return v, nil
}

func (s *server) openapi(w http.ResponseWriter) {
	document, err := s.disco.OpenAPISchema()
	if err != nil {
		serverError(w, err)
	} else {
		ok(w, document)
	}
}
