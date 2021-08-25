package cluster

import (
	"context"
	"fmt"
	"github.com/argoproj-labs/multi-cluster-kubernetes/api/rest"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/homedir"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func NewAddCommand() *cobra.Command {
	var (
		kubeconfig string
		namespace  string
		host       string
	)
	cmd := &cobra.Command{
		Use: "add [CLUSTER_NAME [CONTEXT_NAME]]",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			startingConfig, err := clientcmd.NewDefaultPathOptions().GetStartingConfig()
			if err != nil {
				return err
			}

			clusterName := startingConfig.CurrentContext
			contextName := startingConfig.CurrentContext
			switch len(args) {
			case 0:
			case 1:
				clusterName = args[0]
			case 2:
				clusterName = args[0]
				contextName = args[1]
			default:
				return fmt.Errorf("expected 0, 1 or 2 args")
			}
			kubeContext, ok := startingConfig.Contexts[contextName]
			if !ok {
				return fmt.Errorf("context named \"%s\" not found, you can list contexts with: `kubectl config get-contexts`", contextName)
			}
			user := startingConfig.AuthInfos[kubeContext.AuthInfo]
			if namespace == "" {
				namespace = kubeContext.Namespace
			}
			c, err := clientcmd.NewDefaultClientConfig(*startingConfig, &clientcmd.ConfigOverrides{Context: *kubeContext}).ClientConfig()
			if err != nil {
				return err
			}
			restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				return err
			}
			client := kubernetes.NewForConfigOrDie(restConfig)
			if err := rest.AddConfig(ctx, clusterName, *c, *user, client.CoreV1().Secrets(namespace), rest.WithHost(host)); err != nil {
				return err
			}
			fmt.Printf("cluster %q added\n", clusterName)
			return nil
		},
	}
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace")
	cmd.Flags().StringVar(&host, "host", "", "(optional) the host, e.g. https://kubernetes.default.svc")
	return cmd
}
