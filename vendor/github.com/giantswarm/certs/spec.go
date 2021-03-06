package certs

type Interface interface {
	// SearchCluster searches for secrets containing TLS certs for guest
	// clusters components.
	SearchCluster(clusterID string) (Cluster, error)
	// SearchClusterOperator searches for secrets containing TLS certs for
	// connecting to guest clusters.
	SearchClusterOperator(clusterID string) (ClusterOperator, error)
	// SearchDraining searches for secrets containing TLS certs for
	// draining nodes in guest clusters.
	SearchDraining(clusterID string) (Draining, error)
	// SearchMonitoring searches for secrets containing TLS certs for
	// monitoring guest clusters.
	SearchMonitoring(clusterID string) (Monitoring, error)
}
