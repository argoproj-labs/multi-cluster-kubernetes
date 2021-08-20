package server

import (
	"encoding/json"
	"fmt"
	gorillaschema "github.com/gorilla/schema"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
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
		groupVersion := parts[2]
		list, err := disco.ServerResourcesForGroupVersion(groupVersion)
		done(w, list, err)
	case 4:
		version := parts[2]
		resource := parts[3]
		gvr := schema.GroupVersionResource{Version: version, Resource: resource}
		switch r.Method {
		case "GET":
			list, err := s.clusterList(r, gvr)
			done(w, list, err)
		default:
			nok(w, errors.NewMethodNotSupported(gvr.GroupResource(), r.Method))
		}
	case 5:
		version := parts[2]
		resource := parts[3]
		clusterName, name := split(parts[4])
		gvr := schema.GroupVersionResource{Version: version, Resource: resource}
		switch r.Method {
		case "GET":
			v, err := s.clusterGet(r, clusterName, name, gvr)
			done(w, v, err)
		default:
			nok(w, errors.NewMethodNotSupported(gvr.GroupResource(), r.Method))
		}
	case 6:
		version := parts[2]
		clusterName, namespace := split(parts[4])
		resource := parts[5]
		gvr := schema.GroupVersionResource{Version: version, Resource: resource}
		switch r.Method {
		case "GET":
			if r.URL.Query().Get("watch") == "true" {
				_watch, err := s.watch(r, clusterName, namespace, gvr)
				if err != nil {
					nok(w, err)
				} else {
					stream(w, _watch, clusterName, resource)
				}
			} else {
				list, err := s.list(r, clusterName, namespace, gvr)
				done(w, list, err)
			}
		default:
			nok(w, errors.NewMethodNotSupported(gvr.GroupResource(), r.Method))
		}
	case 7:
		version := parts[2]
		clusterName, namespace := split(parts[4])
		resource := parts[5]
		name := parts[6]
		gvr := schema.GroupVersionResource{Version: version, Resource: resource}
		switch r.Method {
		case "GET":
			v, err := s.get(r, clusterName, namespace, name, gvr)
			done(w, v, err)
		case "PATCH":
			v, err := s.patch(r, clusterName, namespace, name, gvr)
			done(w, v, err)
		default:
			nok(w, errors.NewMethodNotSupported(gvr.GroupResource(), r.Method))
		}
	default:
		nok(w, errors.NewInternalError(fmt.Errorf("unexpected number of path parts %d", len(parts))))
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
		group := parts[2]
		version := parts[3]
		resource := parts[4]
		gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
		switch r.Method {
		case "GET":
			list, err := s.clusterList(r, gvr)
			done(w, list, err)
		default:
			nok(w, errors.NewMethodNotSupported(gvr.GroupResource(), r.Method))
		}
	case 6:
		group := parts[2]
		version := parts[3]
		resource := parts[4]
		clusterName, name := split(parts[5])
		gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
		switch r.Method {
		case "GET":
			v, err := s.clusterGet(r, clusterName, name, gvr)
			done(w, v, err)
		case "PATCH":
			v, err := s.clusterPatch(r, clusterName, name, gvr)
			done(w, v, err)
		default:
			nok(w, errors.NewMethodNotSupported(gvr.GroupResource(), r.Method))
		}
	case 7:
		group := parts[2]
		version := parts[3]
		clusterName, namespace := split(parts[5])
		resource := parts[6]
		gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
		switch r.Method {
		case "GET":
			list, err := s.list(r, clusterName, namespace, gvr)
			done(w, list, err)
		case "POST":
			v, err := s.create(r, clusterName, namespace, gvr)
			done(w, v, err)
		default:
			nok(w, errors.NewMethodNotSupported(gvr.GroupResource(), r.Method))
		}
	case 8:
		group := parts[2]
		version := parts[3]
		clusterName, namespace := split(parts[5])
		resource := parts[6]
		name := parts[7]
		gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
		switch r.Method {
		case "GET":
			v, err := s.get(r, clusterName, namespace, name, gvr)
			done(w, v, err)
		case "PATCH":
			v, err := s.patch(r, clusterName, namespace, name, gvr)
			done(w, v, err)
		default:
			nok(w, errors.NewMethodNotSupported(gvr.GroupResource(), r.Method))
		}
	default:
		nok(w, errors.NewInternalError(fmt.Errorf("unexpected number of path parts %d", len(parts))))
	}
}

func (s *server) create(r *http.Request, clusterName, namespace string, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	fmt.Printf("create %s.%s/%s\n", clusterName, namespace, gvr.String())
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
	v, err := client.Resource(gvr).Namespace(namespace).Create(r.Context(), obj, opts)
	if err != nil {
		return nil, err
	}
	setMetaData(v, clusterName, gvr.Resource)
	return v, nil
}

func (s *server) clusterList(r *http.Request, gvr schema.GroupVersionResource) (*unstructured.UnstructuredList, error) {
	fmt.Printf("cluster list %s\n", gvr.String())
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

func (s *server) list(r *http.Request, clusterName, namespace string, gvr schema.GroupVersionResource) (*unstructured.UnstructuredList, error) {
	fmt.Printf("list %s.%s/%s\n", clusterName, namespace, gvr.String())
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
		setMetaData(&v, clusterName, gvr.Resource)
	}
	return list, nil
}

func (s *server) watch(r *http.Request, clusterName, namespace string, gvr schema.GroupVersionResource) (watch.Interface, error) {
	fmt.Printf("watch %s.%s/%s\n", clusterName, namespace, gvr.String())
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

func setMetaData(v *unstructured.Unstructured, clusterName string, resource string) {
	v.SetClusterName(clusterName)
	if resource == "namespaces" {
		v.SetName(join(v.GetClusterName(), v.GetName()))
	} else if v.GetNamespace() != "" {
		v.SetNamespace(join(v.GetClusterName(), v.GetNamespace()))
	}
}

func (s *server) get(r *http.Request, clusterName, namespace, name string, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	fmt.Printf("get %s.%s/%s/%s\n", clusterName, namespace, name, gvr.String())
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
	setMetaData(v, clusterName, gvr.Resource)
	return v, nil
}

func (s *server) patch(r *http.Request, clusterName, namespace, name string, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	fmt.Printf("patch %s.%s/%s/%s\n", clusterName, namespace, name, gvr.String())
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
	setMetaData(v, clusterName, gvr.Resource)
	return v, nil
}

func (s *server) clusterGet(r *http.Request, clusterName, name string, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	fmt.Printf("get %s.%s/%s\n", clusterName, name, gvr.String())
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
	setMetaData(v, clusterName, gvr.Resource)
	return v, nil
}

func (s *server) clusterPatch(r *http.Request, clusterName, name string, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	fmt.Printf("patch %s/%s/%s\n", clusterName, name, gvr.String())
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
	setMetaData(v, clusterName, gvr.Resource)
	return v, nil
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
