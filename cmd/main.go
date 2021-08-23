package main

import (
	"github.com/argoproj-labs/multi-cluster-kubernetes-api/cmd/cluster"
	"github.com/argoproj-labs/multi-cluster-kubernetes-api/cmd/server"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	cmd := cobra.Command{}
	cmd.AddCommand(cluster.NewCommand())
	cmd.AddCommand(server.NewCommand())
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
