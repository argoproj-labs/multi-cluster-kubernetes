# Multi-Cluster Kubernetes

This is a proof-of-concept to determine if what happens when we amalgamate the APIs from multiple Kubernetes API
servers.

## Install

```
go install github.com/argoproj-labs/multi-cluster-kubernetes-api/cmd/mck
```

Add your current `docker-desktop` as a cluster:

```
mc cluster add default docker-desktop
```


