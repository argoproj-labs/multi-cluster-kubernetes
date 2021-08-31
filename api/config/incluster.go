package config

import (
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var (
	InClusterCluster = &clientcmdapi.Cluster{
		Server:               "https://kubernetes.default.svc",
		CertificateAuthority: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
	}
	InClusterUser = &clientcmdapi.AuthInfo{
		TokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
	}
)
