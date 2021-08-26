package config

import (
	"context"
	"fmt"
	"github.com/argoproj-labs/multi-cluster-kubernetes/api/config"
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
		Use:     "rm [CONFIG_NAME]",
		Aliases: []string{"rm"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			startingConfig, err := clientcmd.NewDefaultPathOptions().GetStartingConfig()
			if err != nil {
				return err
			}

			configName := startingConfig.CurrentContext
			switch len(args) {
			case 0:
			case 1:
				configName = args[0]
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
			if err := config.RemoveConfig(ctx, configName, client.CoreV1().Secrets(namespace)); err != nil {
				return err
			}
			fmt.Printf("config %q removed\n", configName)
			return nil
		},
	}
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace")
	return cmd
}
