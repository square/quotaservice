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
package test

import (
	"testing"
	"github.com/maniksurtani/quotaservice/configs"
	"github.com/maniksurtani/quotaservice/buckets"
	"github.com/maniksurtani/quotaservice/logging"
	"strconv"
	"time"
)

func TestGC(t *testing.T) {
	for impl, factory := range factories {
		logging.Printf("Testing %v", impl)
		cfg := configs.NewDefaultServiceConfig()
		cfg.Namespaces["n"] = configs.NewDefaultNamespaceConfig()
		cfg.Namespaces["n"].DynamicBucketTemplate = configs.NewDefaultBucketConfig()
		// Times out every 5 seconds
		cfg.Namespaces["n"].DynamicBucketTemplate.MaxIdleMillis = 5000
		container := buckets.NewBucketContainer(cfg, factory)

		// No GC should happen here as long as we are in use.
		for i := 0; i < 10; i++ {
			for j := 0; j < 10; j++ {
				bName := strconv.Itoa(j)
				b := container.FindBucket("n", bName)
				if b == nil {
					t.Fatalf("Failed looking for bucket %v on impl %v", bName, impl)
				}

				// Check that the bucket hasn't been GC'd
				if !container.Exists("n", bName) {
					t.Fatalf("Bucket %v was GC'd when it shouldn't have on impl %v", bName, impl)
				}
			}
		}

		// Time out.
		time.Sleep(time.Duration(cfg.Namespaces["n"].DynamicBucketTemplate.MaxIdleMillis) * time.Millisecond * 5)

		// GC should happen here after sleep.
		for i := 0; i < 10; i++ {
			for j := 0; j < 10; j++ {
				bName := strconv.Itoa(j)
				// Check that the bucket has been GC'd
				if container.Exists("n", bName) {
					t.Fatalf("Bucket %v wasn't GC'd when it should have on impl %v", bName, impl)
				}
			}
		}
	}
}
