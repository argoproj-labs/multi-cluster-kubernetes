package main

import (
	cluster "github.com/argoproj-labs/multi-cluster-kubernetes-api/cmd/mc/cluster"
	server "github.com/argoproj-labs/multi-cluster-kubernetes-api/cmd/mc/server"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	cmd := cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}
	cmd.AddCommand(cluster.NewCommand())
	cmd.AddCommand(server.NewCommand())
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
