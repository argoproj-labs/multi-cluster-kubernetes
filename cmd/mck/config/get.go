package config

import (
	"context"
	"fmt"
	"github.com/argoproj-labs/multi-cluster-kubernetes/api/config"
	"k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/yaml"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

func NewGetCommand() *cobra.Command {
	var (
		kubeconfig string
		namespace  string
		raw        bool
	)
	cmd := &cobra.Command{
		Use: "get NAME",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			name := args[0]

			clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}, &clientcmd.ConfigOverrides{})
			restConfig, err := clientConfig.ClientConfig()
			if err != nil {
				return err
			}
			if namespace == "" {
				namespace, _, err = clientConfig.Namespace()
				if err != nil {
					return err
				}
			}

			secretsInterface := kubernetes.NewForConfigOrDie(restConfig).CoreV1().Secrets(namespace)

			c, err := config.New(secretsInterface).Get(ctx, name)
			if err != nil {
				return err
			}

			if !raw {
				api.ShortenConfig(c)
			}
			data, err := yaml.Marshal(c)
			if err != nil {
				return err
			}
			fmt.Println(string(data))

			return nil
		},
	}
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace")
	cmd.Flags().BoolVar(&raw, "raw", false, "raw")
	return cmd
}
