# Amalgamated Kubernetes API

This is a proof-of-concept to determine if what happens when we amalgamate the APIs from multiple Kubernetes API
servers.

Interesting things:

* For namespace scoped resources, the namespace is replaced by "clusterName.namespace".
* For namespaces, the name is replaced by "clusterName.name".
* CRDs need special treatment, no changing their names.
* Some resources (`Role` and `Event`) have the namespace in the object (e.g. `subjects.namespace`
  or `involvedObject.namespace`) and need special treatment.
* Something strange going on with `resourceVersion`.