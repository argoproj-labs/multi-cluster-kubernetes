package config

import (
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"os"
)

// InCluster is a reserved name for the in-cluster configuration.
const InCluster = "@in-cluster"

var (
	InClusterCluster = &clientcmdapi.Cluster{
		Server:               "https://kubernetes.default.svc",
		CertificateAuthority: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
	}
	InClusterUser = &clientcmdapi.AuthInfo{
		TokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
	}
)

func InClusterConfig() clientcmdapi.Config {
	namespace, _ := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	return clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{InCluster: InClusterCluster},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			InCluster: InClusterUser,
		},
		Contexts: map[string]*clientcmdapi.Context{
			InCluster: {
				Cluster:   InCluster,
				AuthInfo:  InCluster,
				Namespace: string(namespace),
			},
		},
		CurrentContext: InCluster,
	}
}
