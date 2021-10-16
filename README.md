# Multi-Cluster Kubernetes

A library to support writing Kubernetes applications that need to connect to multiple clusters.

## Features

* Cluster credentials storage - loading and saving KUBECONFIG with secrets
* Multi-cluster/namespace resource ownership patterns - labelling resources to identify the owner
* Multi-cluster implementations of:
    * `MetaNamespaceKeyFunc` and friends
    * `kuberentes.Interface`
    * `dynamic.Interface`
    * `cache.SharedIndexerInformer`
