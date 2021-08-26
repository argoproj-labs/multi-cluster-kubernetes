package main

import (
	"github.com/argoproj-labs/multi-cluster-kubernetes/cmd/mck/config"
	server "github.com/argoproj-labs/multi-cluster-kubernetes/cmd/mck/server"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	cmd := cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}
	cmd.AddCommand(config.NewCommand())
	cmd.AddCommand(server.NewCommand())
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
