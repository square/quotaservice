// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package metrics

type Metrics interface {
	GlobalDefaultBucketMetrics() BucketMetrics
	Namespaces() map[string]NamespaceMetrics
	Reset()
}

type NamespaceMetrics interface {
	Name() string
	NumBuckets() int
	NumDynamicBuckets() int
	DefaultBucketMetrics() BucketMetrics
	Buckets() map[string]BucketMetrics
}

type BucketMetrics interface {
	Name() string
	// TODO(manik)
}
