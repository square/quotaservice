// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package admin

import (
	"errors"

	"github.com/maniksurtani/quotaservice/config"
	pb "github.com/maniksurtani/quotaservice/protos/config"
	"github.com/maniksurtani/quotaservice/stats"
)

type MockAdministrable struct {
	cfg    *pb.ServiceConfig
	errors bool
}

func NewMockErrorAdministrable() *MockAdministrable {
	return &MockAdministrable{config.NewDefaultServiceConfig(), true}
}

func NewMockAdministrable() *MockAdministrable {
	return &MockAdministrable{config.NewDefaultServiceConfig(), false}
}

func (m *MockAdministrable) Configs() *pb.ServiceConfig {
	return m.cfg
}

func (m *MockAdministrable) UpdateConfig(config *pb.ServiceConfig, user string) error {
	if m.errors {
		return errors.New("UpdateConfig")
	}

	return nil
}

func (m *MockAdministrable) DeleteBucket(namespace string, name string) error {
	if m.errors {
		return errors.New("DeleteBucket")
	}

	return nil
}

func (m *MockAdministrable) AddBucket(namespace string, b *pb.BucketConfig) error {
	if m.errors {
		return errors.New("AddBucket")
	}

	return nil
}

func (m *MockAdministrable) UpdateBucket(namespace string, b *pb.BucketConfig) error {
	if m.errors {
		return errors.New("UpdateBucket")
	}

	return nil
}

func (m *MockAdministrable) DeleteNamespace(namespace string) error {
	if m.errors {
		return errors.New("DeleteNamespace")
	}

	return nil
}

func (m *MockAdministrable) AddNamespace(n *pb.NamespaceConfig) error {
	if m.errors {
		return errors.New("AddNamespace")
	}

	return nil
}

func (m *MockAdministrable) UpdateNamespace(n *pb.NamespaceConfig) error {
	if m.errors {
		return errors.New("UpdateNamespace")
	}

	return nil
}

func (m *MockAdministrable) TopDynamicHits(namespace string) []*stats.BucketScore {
	if m.errors {
		return nil
	}

	return make([]*stats.BucketScore, 0)
}

func (m *MockAdministrable) TopDynamicMisses(namespace string) []*stats.BucketScore {
	if m.errors {
		return nil
	}

	return make([]*stats.BucketScore, 0)
}

func (m *MockAdministrable) DynamicBucketStats(namespace, bucket string) *stats.BucketScores {
	if m.errors {
		return nil
	}

	return &stats.BucketScores{0, 0}
}

func (m *MockAdministrable) HistoricalConfigs() ([]*pb.ServiceConfig, error) {
	if m.errors {
		return nil, errors.New("HistoricalConfigs")
	}

	return make([]*pb.ServiceConfig, 1), nil
}
