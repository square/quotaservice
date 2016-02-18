package clustering

// Clustering is the interface you must implement if you intend to run the quota service in a
// clustered environment. You would typically back this with a hook in to your organization's
// service discovery mechanism, such as Zookeeper.
type Clustering interface {
	// CurrentNodeName returns the current node name, as a string.
	CurrentNodeName() string
	// Members returns a slice of node names that form a cluster.
	Members() []string
	// MembershipChangeNotificationChannel returns a channel that is used to notify listeners of a
	// membership change.
	MembershipChangeNotificationChannel() chan bool
}
