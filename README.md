# Multi-cluster Kubernetes API

This is a proof-of-concept to determine if what happens when we amalgamate the APIs from multiple Kubernetes API
servers.

Interesting things:

* For namespace scoped resources, the namespace is replaced by `clusterName.namespace`. This seems to work without
  issues.
* For cluster scoped resources, the name is replaced by `clusterName.name`.
* Unless they are `CustomResourceDefinition`.
* Some resources (`Role` and `Event`) have the namespace in the object (e.g. `subjects.namespace`
  or `involvedObject.namespace`) and need special treatment.
* The API exposes some special content types, but we may not need to understand them (
  e.g. `Accept: application/json;as=Table;v=v1;g=meta.k8s.io,application/json;as=Table;v=v1beta1;g=meta.k8s.io,application/json`)
  because it seems to be fine with JSON.
