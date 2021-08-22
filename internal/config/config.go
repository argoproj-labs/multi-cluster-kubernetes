package config

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

func NewRestConfigOrDie(kubeconfig string, namespace *string) *rest.Config {
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{})
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}
	if namespace != nil && *namespace == "" {
		s, _, err := kubeConfig.Namespace()
		if err != nil {
			log.Fatal(err)
		}
		*namespace = s
	}
	return config
}
