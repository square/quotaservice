package clustering

type Clustering interface {
	// Returns the current node name
	CurrentNodeName() string
	// Returns a slice of node names that form a cluster.
	Members() []string
	// Returns a channel that is used to notify listeners of a membership change.
	MembershipChangeNotificationChannel() chan bool
}
