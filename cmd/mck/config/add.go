package config

import (
	"context"
	"fmt"
	"github.com/argoproj-labs/multi-cluster-kubernetes/api/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

func NewAddCommand() *cobra.Command {
	var (
		kubeconfig      string
		labels          []string
		namespace       string
		configNamespace string
		host            string
	)
	const undefined = "-"
	cmd := &cobra.Command{
		Use: "add [CONFIG_NAME [CONTEXT_NAME]]",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			startingConfig, err := clientcmd.NewDefaultPathOptions().GetStartingConfig()
			if err != nil {
				return err
			}

			configName := startingConfig.CurrentContext
			contextName := startingConfig.CurrentContext
			switch len(args) {
			case 0:
			case 1:
				configName = args[0]
			case 2:
				configName = args[0]
				contextName = args[1]
			default:
				return fmt.Errorf("expected 0, 1 or 2 args")
			}
			kubeContext, ok := startingConfig.Contexts[contextName]
			if !ok {
				return fmt.Errorf("context named \"%s\" not found, you can list contexts with: `kubectl config get-contexts`", contextName)
			}
			if namespace == "" {
				namespace = kubeContext.Namespace
			}
			restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				return err
			}
			client := kubernetes.NewForConfigOrDie(restConfig)

			var opts []interface{}
			if host != undefined {
				opts = append(opts, config.WithServer(host))
			}
			if configNamespace != undefined {
				opts = append(opts, config.WithNamespace(configNamespace))
			}
			for _, label := range labels {
				parts := strings.SplitN(label, "=", 2)
				opts = append(opts, config.WithLabel(parts[0], parts[1]))
			}
			if c, err := config.AddConfig(
				ctx,
				configName,
				startingConfig,
				contextName,
				client.CoreV1().Secrets(namespace),
				opts...,
			); err != nil {
				return err
			} else {
				kubeConfig := c.Contexts[c.CurrentContext]
				fmt.Printf("config %q (%s/%s) added\n", configName, kubeConfig.Cluster, kubeContext.Namespace)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace to add config too (default: current namespace)")
	cmd.Flags().StringVar(&configNamespace, "config-namespace", undefined, "namespace to add config too (default: the context's namespace)")
	cmd.Flags().StringArrayVarP(&labels, "labels", "l", nil, "labels to add config")
	cmd.Flags().StringVar(&host, "host", undefined, "the host, e.g. https://kubernetes.default.svc (default: the context's cluster's host)")
	return cmd
}
