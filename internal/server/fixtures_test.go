package server

import (
	"context"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

func RestConfig() *rest.Config {
	restConfig, err := clientcmd.BuildConfigFromFlags("", "../../KubeConfig.yaml")
	if err != nil {
		panic(err)
	}
	return restConfig
}

func Setup() func() {
	namespace := "default"
	restConfig, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		panic(err)
	}
	shutdown, err := New(restConfig, namespace)
	if err != nil {
		panic(err)
	}
	return func() { _ = shutdown(context.Background()) }
}
