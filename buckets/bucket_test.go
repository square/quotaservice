/*
 *   Copyright 2016 Manik Surtani
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

// Package buckets defines interfaces for abstractions of token buckets.
package buckets

import (
	"testing"
	"github.com/maniksurtani/quotaservice/configs"
	"time"
)

// Mock objects
type mockBucket struct {
	namespace, bucketName string
}

func (b *mockBucket) Take(numTokens int64, maxWaitTime time.Duration) (waitTime time.Duration) {
	return 0
}
func (b *mockBucket) Config() *configs.BucketConfig {
	return nil
}
func (b *mockBucket) ActivityDetected() bool {
	return true
}
func (b *mockBucket) ReportActivity() {}
func (b *mockBucket) Dynamic() bool {return false}

type mockBucketFactory struct{}

func (bf mockBucketFactory) Init(cfg *configs.ServiceConfig) {}
func (bf mockBucketFactory) NewBucket(namespace string, bucketName string, cfg *configs.BucketConfig) Bucket {
	return &mockBucket{namespace: namespace, bucketName: bucketName}
}

var cfg = func() *configs.ServiceConfig {
	c := configs.NewDefaultServiceConfig()
	c.GlobalDefaultBucket = configs.NewDefaultBucketConfig()
	c.Namespaces["x"] = configs.NewDefaultNamespaceConfig()
	c.Namespaces["x"].DefaultBucket = configs.NewDefaultBucketConfig()
	c.Namespaces["x"].Buckets["a"] = configs.NewDefaultBucketConfig()
	c.Namespaces["y"] = configs.NewDefaultNamespaceConfig()
	c.Namespaces["y"].DynamicBucketTemplate = configs.NewDefaultBucketConfig()
	c.Namespaces["y"].Buckets["a"] = configs.NewDefaultBucketConfig()
	return c
}()

var container = NewBucketContainer(cfg, &mockBucketFactory{})

func TestFallbackToGlobalDefaultBucket(t *testing.T) {
	b := container.FindBucket("nonexistent_namespace", "nonexistent_bucket")
	if b == nil {
		t.Fatal("Should fall back to default bucket.")
	}

	if b != container.defaultBucket {
		t.Fatal("Should fall back to default bucket.")
	}
}

func TestFallbackToDefaultBucket(t *testing.T) {
	b := container.FindBucket("x", "nonexistent_bucket")
	if b == nil {
		t.Fatal("Should fall back to default bucket.")
	}

	if b != container.namespaces["x"].defaultBucket {
		t.Fatal("Should fall back to default bucket.")
	}
}

func TestDynamicBucket(t *testing.T) {
	b := container.FindBucket("y", "new")
	if b == nil {
		t.Fatal("Should create new bucket.")
	}

	if b != container.namespaces["y"].buckets["new"] {
		t.Fatal("Should create new bucket.")
	}
}

func TestBucketNamespaces(t *testing.T) {
	bx := container.FindBucket("x", "a")
	if bx == nil {
		t.Fatal("Should create new bucket.")
	}

	if bx != container.namespaces["x"].buckets["a"] {
		t.Fatal("Should create new bucket.")
	}

	by := container.FindBucket("y", "a")
	if by == nil {
		t.Fatal("Should create new bucket.")
	}

	if by != container.namespaces["y"].buckets["a"] {
		t.Fatal("Should create new bucket.")
	}

	if by == bx {
		t.Fatal("Buckets in different namespaces should be different instances.")
	}
}
