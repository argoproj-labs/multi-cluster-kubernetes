# Multi-Cluster Kubernetes API

This is a proof-of-concept to determine if what happens when we amalgamate the APIs from multiple Kubernetes API
servers.

* The API exposes some special content types, but we may not need to understand them (
  e.g. `Accept: application/json;as=Table;v=v1;g=meta.k8s.io,application/json;as=Table;v=v1beta1;g=meta.k8s.io,application/json`)
  because it seems to be fine with JSON.