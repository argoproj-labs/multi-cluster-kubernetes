package labels

const (
	KeyClusterName      = "multi-cluster.argoproj.io/cluster-name"       // which cluster this resource is in
	KeyOwnerClusterName = "multi-cluster.argoproj.io/owner-cluster-name" // which cluster contains the owner
	KeyOwnerNamespace   = "multi-cluster.argoproj.io/owner-namespace"    // which namespace contains the owner
	KeyOwnerName        = "multi-cluster.argoproj.io/owner-name"         // owner name
	KeyRestConfig       = "rest-config.multi-cluster.argoproj.io"
)
