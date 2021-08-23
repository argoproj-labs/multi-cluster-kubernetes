package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/argoproj-labs/multi-cluster-kubernetes-api/api"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func NewAddCommand() *cobra.Command {
	var (
		kubeconfig string
		namespace  string
	)
	cmd := &cobra.Command{
		Use: "add CLUSTER_NAME CONTEXT_NAME",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("expected 2 args")
			}
			ctx := context.Background()
			cluster := args[0]
			contextName := args[1]
			startingConfig, err := clientcmd.NewDefaultPathOptions().GetStartingConfig()
			if err != nil {
				return err
			}
			kubeContext, ok := startingConfig.Contexts[contextName]
			if !ok {
				log.Fatalf("context named \"%s\" not found, you can list contexts with: `kubectl config get-contexts`", contextName)
			}
			user := startingConfig.AuthInfos[kubeContext.AuthInfo]
			c, err := clientcmd.NewDefaultClientConfig(*startingConfig, &clientcmd.ConfigOverrides{Context: *kubeContext}).ClientConfig()
			if err != nil {
				return err
			}
			data, err := json.Marshal(api.NewConfig(*c, *user))
			if err != nil {
				return err
			}
			data, err = json.Marshal(map[string]map[string]string{"stringData": {cluster: string(data)}})
			if err != nil {
				return err
			}
			restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				return err
			}
			_, err = kubernetes.NewForConfigOrDie(restConfig).CoreV1().Secrets(namespace).
				Patch(ctx, "clusters", types.MergePatchType, data, metav1.PatchOptions{})
			if err != nil {
				return err
			}
			fmt.Printf("added cluster %q from context %q\n", cluster, contextName)
			return nil
		},
	}
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace")
	return cmd
}
