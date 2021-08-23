package server

import (
	"context"
	"github.com/argoproj-labs/multi-cluster-kubernetes/internal/server"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"os/signal"
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
			ctx := context.Background()
			restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				return err
			}
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, os.Interrupt)
			shutdown, err := server.New(restConfig, namespace)
			if err != nil {
				return err
			}
			<-stop
			return shutdown(ctx)
		},
	}
	cmd.Flags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace")
	return cmd
}
