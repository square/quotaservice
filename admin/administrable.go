// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package admin

import (
	pb "github.com/square/quotaservice/protos/config"
	"github.com/square/quotaservice/stats"
)

// Administrable defines something that can be administered via this package.
type Administrable interface {
	Configs() *pb.ServiceConfig
	HistoricalConfigs() ([]*pb.ServiceConfig, error)

	UpdateConfig(*pb.ServiceConfig, string) error

	DeleteBucket(string, string, string) error
	AddBucket(string, *pb.BucketConfig, string) error
	UpdateBucket(string, *pb.BucketConfig, string) error

	DeleteNamespace(string, string) error
	AddNamespace(*pb.NamespaceConfig, string) error
	UpdateNamespace(*pb.NamespaceConfig, string) error

	TopDynamicHits(string) []*stats.BucketScore
	TopDynamicMisses(string) []*stats.BucketScore
	DynamicBucketStats(string, string) *stats.BucketScores
}
