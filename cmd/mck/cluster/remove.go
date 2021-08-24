package cluster

import (
	"context"
	"fmt"
	"github.com/argoproj-labs/multi-cluster-kubernetes/api"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/homedir"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func NewRemoveCommand() *cobra.Command {
	var (
		kubeconfig string
		namespace  string
	)
	cmd := &cobra.Command{
		Use:     "rm [CLUSTER_NAME]",
		Aliases: []string{"rm"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			startingConfig, err := clientcmd.NewDefaultPathOptions().GetStartingConfig()
			if err != nil {
				return err
			}

			clusterName := startingConfig.CurrentContext
			switch len(args) {
			case 0:
			case 1:
				clusterName = args[0]
			default:
				return fmt.Errorf("expected 0 or 1 args")
			}
			if c, ok := startingConfig.Contexts[startingConfig.CurrentContext]; ok && namespace == "" {
				namespace = c.Namespace
			}

			restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				return err
			}
			client := kubernetes.NewForConfigOrDie(restConfig)
			if err := api.RemoveCluster(ctx, clusterName, client.CoreV1().Secrets(namespace)); err != nil {
				return err
			}
			fmt.Printf("cluster %q removed\n", clusterName)
			return nil
		},
	}
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace")
	return cmd
}
