package server

import (
	config "github.com/argoproj-labs/multi-cluster-kubernetes-api/internal/config"
	"github.com/argoproj-labs/multi-cluster-kubernetes-api/internal/server"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

func NewCommand() *cobra.Command {
	var (
		kubeconfig string
		namespace  string
	)
	cmd := &cobra.Command{
		Use: "server",
		RunE: func(cmd *cobra.Command, args []string) error {
			restConfig := config.NewRestConfigOrDie(kubeconfig, &namespace)
			return server.New(restConfig, namespace)
		},
	}
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace")
	return cmd
}
