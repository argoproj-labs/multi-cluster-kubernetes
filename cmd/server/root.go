package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj-labs/amalgamated-kubernetes-api/clusters"
	"github.com/spf13/cobra"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

import gorillaschema "github.com/gorilla/schema"

var decoder = gorillaschema.NewDecoder()

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

func serverError(w http.ResponseWriter, err error) {
	fmt.Printf("500\n")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)
	_ = json.NewEncoder(io.MultiWriter(os.Stderr, w)).Encode(metav1.Status{
		Status:  "Failure",
		Message: err.Error(),
	})
}

func NewCommand() *cobra.Command {
	var kubeconfig string
	var namespace string
	cmd := &cobra.Command{
		Use: "server",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				panic(err)
			}
			secret, err := kubernetes.NewForConfigOrDie(config).CoreV1().Secrets(namespace).Get(context.Background(), "clusters", metav1.GetOptions{})
			if err != nil {
				return err
			}
			clients := make(map[string]dynamic.Interface)
			for clusterName, data := range secret.Data {
				c := &clusters.Config{}
				if err := json.Unmarshal(data, c); err != nil {
					return err
				}
				clients[clusterName] = dynamic.NewForConfigOrDie(c.RestConfig())
			}
			disco := discovery.NewDiscoveryClientForConfigOrDie(config)
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				parts := strings.Split(r.URL.Path, "/")
				query := r.URL.Query()
				fmt.Printf("%s %q (%d) %q\n", r.Method, parts, len(parts), query)
				switch parts[1] {
				case "apis":
					switch len(parts) {
					case 2:
						groups, err := disco.ServerGroups()
						done(w, groups, err)
					case 4:
						groupVersion, err := disco.ServerResourcesForGroupVersion(parts[2] + "/" + parts[3])
						done(w, groupVersion, err)
					case 7:
						switch r.Method {
						case "GET":
							clusterName, namespace := split(parts[5])
							list(w, r, parts[2], parts[3], clusterName, namespace, parts[6], clients[clusterName])
						case "POST":
							clusterName, namespace := split(parts[5])
							create(w, r, parts[2], parts[3], clusterName, namespace, parts[6], clients[clusterName])
						default:
							serverError(w, fmt.Errorf("unexpected method"))
						}
					case 8:
						clusterName, namespace := split(parts[5])
						get(w, r, parts[2], parts[3], clusterName, namespace, parts[6], parts[7], clients[clusterName])
					default:
						serverError(w, fmt.Errorf("unexpected number of path parts"))
					}
				default:
					switch len(parts) {
					case 2:
						ok(w, metav1.APIVersions{TypeMeta: metav1.TypeMeta{Kind: "APIVersions"}, Versions: []string{}})
					case 3:
						list, err := disco.ServerResourcesForGroupVersion(parts[2])
						done(w, list, err)
					case 6:
						clusterName, namespace := split(parts[4])
						list(w, r, "", parts[2], clusterName, namespace, parts[5], clients[clusterName])
					case 7:
						clusterName, namespace := split(parts[4])
						get(w, r, "", parts[2], clusterName, namespace, parts[5], parts[6], clients[clusterName])
					default:
						serverError(w, fmt.Errorf("unexpected number of path parts"))
					}
				}
			})
			server := &http.Server{Addr: ":2473"}
			fmt.Printf("starting on %q\n", server.Addr)
			return server.ListenAndServe()
		},
	}
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace")
	return cmd
}

func create(w http.ResponseWriter, r *http.Request, group, version, clusterName, namespace, resource string, client dynamic.Interface) {
	opts := metav1.CreateOptions{}
	_ = decoder.Decode(&opts, r.URL.Query())
	obj := &unstructured.Unstructured{}
	err := json.NewDecoder(r.Body).Decode(obj)
	if err != nil {
		serverError(w, err)
		return
	}
	obj.SetNamespace(namespace)
	v, err := client.Resource(schema.GroupVersionResource{Group: group, Version: version, Resource: resource}).Namespace(namespace).Create(r.Context(), obj, opts)
	if err != nil {
		serverError(w, err)
	} else {
		v.SetNamespace(join(clusterName, namespace))
		ok(w, v)
	}
}

func join(clusterName, namespace string) string {
	return fmt.Sprintf("%s.%s", clusterName, namespace)
}

func split(s string) (string, string) {
	parts := strings.Split(s, ".")
	return parts[0], parts[1]
}

func list(w http.ResponseWriter, r *http.Request, group, version, clusterName, namespace, resource string, client dynamic.Interface) {
	opts := metav1.ListOptions{}
	_ = decoder.Decode(&opts, r.URL.Query())
	v, err := client.Resource(schema.GroupVersionResource{Group: group, Version: version, Resource: resource}).Namespace(namespace).List(r.Context(), opts)
	if err != nil {
		serverError(w, err)
	} else {
		for _, item := range v.Items {
			item.SetNamespace(join(clusterName, namespace))
		}
		ok(w, v)
	}
}

func get(w http.ResponseWriter, r *http.Request, group, version, clusterName, namespace, resource, name string, client dynamic.Interface) {
	opts := metav1.GetOptions{}
	_ = decoder.Decode(&opts, r.URL.Query())
	v, err := client.Resource(schema.GroupVersionResource{Group: group, Version: version, Resource: resource}).Namespace(namespace).Get(r.Context(), name, opts)
	if err != nil {
		serverError(w, err)
	} else {
		v.SetNamespace(join(clusterName, namespace))
		ok(w, v)
	}
}
