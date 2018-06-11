// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package admin

import (
	"github.com/square/quotaservice/config"
	pb "github.com/square/quotaservice/protos/config"
	"github.com/square/quotaservice/stats"
)

// Administrable defines something that can be administered via this package.
type Administrable interface {
	Configs() *ConfigAndHash
	HistoricalConfigs() ([]*ConfigAndHash, error)

	UpdateConfig(newConfig *pb.ServiceConfig, ctx *Context) error

	DeleteBucket(namespace, bucketName string, ctx *Context) error
	AddBucket(namespace string, bucket *pb.BucketConfig, ctx *Context) error
	UpdateBucket(namespace string, bucket *pb.BucketConfig, ctx *Context) error

	DeleteNamespace(namespace string, ctx *Context) error
	AddNamespace(namespace *pb.NamespaceConfig, ctx *Context) error
	UpdateNamespace(namespace *pb.NamespaceConfig, ctx *Context) error

	TopDynamicHits(namespace string) []*stats.BucketScore
	TopDynamicMisses(namespace string) []*stats.BucketScore
	DynamicBucketStats(namespace string, bucketName string) *stats.BucketScores
}

// Context wraps all necessary components of a config update made via Administrable.
type Context struct {
	User    string
	OldHash string
}

// ConfigAndHash is a struct that contains a pointer to the ServiceConfig and its hash. The latter is used in optimistic
// version checks and when making changes via Administrable, the hash of the ServiceConfig being modified should be
// included. ConfigAndHash can be sorted.
type ConfigAndHash struct {
	*pb.ServiceConfig
	Hash string `json:"hash,omitempty" yaml:"hash"`
}

func NewConfigAndHash(cfg *pb.ServiceConfig) *ConfigAndHash {
	return &ConfigAndHash{cfg, config.HashConfig(cfg)}
}

func NewContext(user, oldHash string) *Context {
	return &Context{user, oldHash}
}

const (
	TODO = "TODO"
)
