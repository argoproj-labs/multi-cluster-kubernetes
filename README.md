# Multi-Cluster Kubernetes

This is a proof-of-concept to determine if what happens when we amalgamate the APIs from multiple Kubernetes API
servers.

## Install

```
go install github.com/argoproj-labs/multi-cluster-kubernetes/cmd/mck@v0.0.1
```

Add your current context as a cluster:

```
mck cluster add
```

## Fetaurse

* Cluster credentials storage - saving KUBECONFIG in secrets
* Cross-cluster/namespace resource ownership - labelling resources to identify the owner


