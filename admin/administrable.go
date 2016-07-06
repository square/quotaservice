// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package admin

import (
	pb "github.com/maniksurtani/quotaservice/protos/config"
)

// Administrable defines something that can be administered via this package.
type Administrable interface {
	Configs() *pb.ServiceConfig

	DeleteBucket(namespace, name string) error
	AddBucket(namespace string, b *pb.BucketConfig) error
	UpdateBucket(namespace string, b *pb.BucketConfig) error

	DeleteNamespace(namespace string) error
	AddNamespace(n *pb.NamespaceConfig) error
	UpdateNamespace(n *pb.NamespaceConfig) error
}
