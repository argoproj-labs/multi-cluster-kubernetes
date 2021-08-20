package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj-labs/amalgamated-kubernetes-api/clusters"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"net/http"
	"path/filepath"
	"strings"
)

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
			server := server{
				Server:  http.Server{Addr: ":2473"},
				disco:   disco,
				clients: clients,
			}
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				parts := strings.Split(r.URL.Path, "/")
				query := r.URL.Query()
				fmt.Printf("%s %q (%d) %q\n", r.Method, parts, len(parts), query)
				switch parts[1] {
				case "api":
					server.api(w, r, parts, disco)
				case "apis":
					server.apis(w, r, parts, disco)
				case "openapi":
					server.openapi(w)
				default:
					serverError(w, fmt.Errorf("unknown %q", parts[1]))
				}
			})
			fmt.Printf("starting on %q\n", server.Addr)
			return server.ListenAndServe()
		},
	}
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace")
	return cmd
}
