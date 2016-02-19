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
package redis

import (
	"testing"
	"github.com/maniksurtani/quotaservice/configs"
)

func TestRedisBucket(t *testing.T) {
	bf := NewBucketFactory()
	// Test that the Redis bucket factory actually doesn't need any more config
	bf.Init(nil)
	cfg := configs.NewDefaultBucketConfig()
	b := bf.NewBucket("ns", "n", cfg)
	if b == nil {
		t.Fatal("Bucket should not be nil")
	}

	if b.GetConfig() != cfg {
		t.Fatal("Bucket cfg should be the same as passed in")
	}
}

