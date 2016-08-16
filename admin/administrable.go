// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package admin

import (
	pb "github.com/maniksurtani/quotaservice/protos/config"
)

// Administrable defines something that can be administered via this package.
type Administrable interface {
	Configs() *pb.ServiceConfig

	UpdateConfig(*pb.ServiceConfig) error

	DeleteBucket(string, string) error
	AddBucket(string, *pb.BucketConfig) error
	UpdateBucket(string, *pb.BucketConfig) error

	DeleteNamespace(string) error
	AddNamespace(*pb.NamespaceConfig) error
	UpdateNamespace(*pb.NamespaceConfig) error
}
