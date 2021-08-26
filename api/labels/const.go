package labels

const (
	KeyConfigName       = "multi-cluster.argoproj.io/config-name"
	KeyOwnerClusterName = "multi-cluster.argoproj.io/owner-cluster-name" // which cluster contains the owner
	KeyOwnerNamespace   = "multi-cluster.argoproj.io/owner-namespace"    // which namespace contains the owner
	KeyOwnerName        = "multi-cluster.argoproj.io/owner-name"         // owner name
)
