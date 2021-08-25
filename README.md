# Multi-Cluster Kubernetes

A library to support writing Kubernetes applications that need to connect to multiple clusters.

## Features

* Cluster credentials storage - loading and saving KUBECONFIG with secrets
* CLI to manage those secrets
* Multi-cluster/namespace resource ownership patterns - labelling resources to identify the owner
* Multi-cluster implementations of:
    * `MetaNamespaceKeyFunc` and friends
    * `kuberentes.Interface`
    * `dynamic.Interface`
    * `cache.SharedIndexerInformer`

## CLI Install

```
go install github.com/argoproj-labs/multi-cluster-kubernetes/cmd/mck@v0.0.1
```

Add your current context as a cluster:

```
mck cluster add
```

