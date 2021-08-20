package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj-labs/amalgamated-kubernetes-api/clusters"
	"github.com/spf13/cobra"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		fmt.Printf("500\n")
		w.WriteHeader(500)
		_ = json.NewEncoder(io.MultiWriter(os.Stderr, w)).Encode(metav1.Status{
			Status:  "Failure",
			Message: err.Error(),
		})
	} else {
		fmt.Printf("200\n")
		w.WriteHeader(200)
		_ = json.NewEncoder(io.MultiWriter(os.Stdout, w)).Encode(v)
	}
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
				fmt.Printf("%q %q\n", parts, query)
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
						clusterName, namespace := split(parts[5])
						list(w, r, parts[2], parts[3], namespace, parts[6], clients[clusterName])
					default:
						done(w, nil, fmt.Errorf("unexpected number of path parts"))
					}
				default:
					switch len(parts) {
					case 2: // /api
						v := metav1.APIVersions{TypeMeta: metav1.TypeMeta{Kind: "APIVersions"}, Versions: []string{}}
						done(w, v, nil)
					case 3: // api/version -- legacy, usually v1
						list, err := disco.ServerResourcesForGroupVersion(parts[2])
						done(w, list, err)
					case 6:
						clusterName, namespace := split(parts[5])
						list(w, r, "", parts[2], namespace, parts[5], clients[clusterName])
					case 7:
						clusterName, namespace := split(parts[5])
						get(w, r, "", parts[2], namespace, parts[5], parts[6], clients[clusterName])
					default:
						done(w, nil, fmt.Errorf("unexpected number of path parts"))
					}
				}
			})
			server := &http.Server{
				Addr: ":8080",
			}
			fmt.Printf("starting on %q\n", server.Addr)
			return server.ListenAndServe()
		},
	}
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace")
	return cmd
}

func split(s string) (string, string) {
	parts := strings.Split(s, ".")
	return parts[0], parts[1]
}

func list(w http.ResponseWriter, r *http.Request, group, version, namespace, resource string, client dynamic.Interface) {
	opts := metav1.ListOptions{}
	_ = decoder.Decode(&opts, r.URL.Query())
	list, err := client.Resource(schema.GroupVersionResource{Group: group, Version: version, Resource: resource}).Namespace(namespace).List(r.Context(), opts)
	done(w, list, err)
}

func get(w http.ResponseWriter, r *http.Request, group, version, namespace, resource, name string, client dynamic.Interface) {
	opts := metav1.GetOptions{}
	_ = decoder.Decode(&opts, r.URL.Query())
	v, err := client.Resource(schema.GroupVersionResource{Group: group, Version: version, Resource: resource}).Namespace(namespace).Get(r.Context(), name, opts)
	done(w, v, err)
}
