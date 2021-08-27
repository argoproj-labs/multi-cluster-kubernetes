package labels

const (
	KeyKubeConfig     = "multi-cluster.argoproj.io/kubeconfig"
	KeyCluster        = "multi-cluster.argoproj.io/cluster"
	KeyOwnerCluster   = "multi-cluster.argoproj.io/owner-cluster"   // which cluster contains the owner
	KeyOwnerNamespace = "multi-cluster.argoproj.io/owner-namespace" // which namespace contains the owner
	KeyOwnerName      = "multi-cluster.argoproj.io/owner-name"      // owner name
)
