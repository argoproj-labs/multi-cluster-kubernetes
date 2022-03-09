package main

import (
	"github.com/argoproj-labs/multi-cluster-kubernetes/cmd/mck/config"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	cmd := cobra.Command{
		Use: "mck",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}
	cmd.AddCommand(config.NewCommand())
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
