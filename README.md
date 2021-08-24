# Multi-Cluster Kubernetes

This is a proof-of-concept to determine if what happens when we amalgamate the APIs from multiple Kubernetes API
servers.

## Install

```
go install github.com/argoproj-labs/multi-cluster-kubernetes/cmd/mck@v0.0.1
```

Add your current `docker-desktop` as a cluster:

```
mck cluster add docker-desktop docker-desktop
```


