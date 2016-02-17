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

package configs
import (
	"testing"
	"github.com/maniksurtani/quotaservice/test"
)

func TestConfig(t *testing.T) {
	yaml := `admin_port: 1234
metrics_enabled: false
filler_frequency_millis: 12345
namespaces:
  no_default_no_dynamic:
    buckets:
      one:
        fill_rate: 321
        wait_timeout_millis: 9999
        max_idle_millis: 20000
      with_defaults:
        size: 100
  only_dynamic:
    dynamic_bucket_template:
      fill_rate: 999
      wait_timeout_millis: 8888
      max_idle_millis: 30000
  only_default:
    default_bucket:
      fill_rate: 800
      wait_timeout_millis: 7777
      max_idle_millis: 40000
`

	cfg := readConfigFromBytes([]byte(yaml))
	if cfg.AdminPort != 1234 {
		t.Fatal("Expected admin_port to be 1234")
	}

	if cfg.MetricsEnabled {
		t.Fatal("Metrics should not be enabled")
	}

	if cfg.FillerFrequencyMillis != 12345 {
		t.Fatal("Expected filler_frequency_millis to be 12345")
	}

	if cfg.GlobalDefaultBucket != nil {
		t.Fatal("Did not configure a global default bucket")
	}

	if len(cfg.Namespaces) != 3 {
		t.Fatal("Expected 3 namespaces")
	}

	namespace := "no_default_no_dynamic"
	ns := cfg.Namespaces[namespace]

	assertNamespace(t, namespace, ns, 2, false, false)
	assertBucket(t, ns.Buckets["one"], 100, 321, 9999, 20000)
	assertBucket(t, ns.Buckets["with_defaults"], 100, 50, 1000, -1)

	namespace = "only_dynamic"
	ns = cfg.Namespaces[namespace]

	assertNamespace(t, namespace, ns, 0, false, true)
	assertBucket(t, ns.DynamicBucketTemplate, 100, 999, 8888, 30000)
	if ns.MaxDynamicBuckets != 0 {
		t.Fatalf("Expecting max_dynamic_buckets to be 0; was %v", ns.MaxDynamicBuckets)
	}

	namespace = "only_default"
	ns = cfg.Namespaces[namespace]

	assertNamespace(t, namespace, ns, 0, true, false)
	assertBucket(t, ns.DefaultBucket, 100, 800, 7777, 40000)
}

func assertNamespace(t *testing.T, namespace string, ns *NamespaceConfig, numBuckets int, expectDefault, expectDynamic bool) {
	if !expectDefault && ns.DefaultBucket != nil {
		t.Fatalf("Did not configure a default bucket for namespace %v", namespace)
	}

	if !expectDynamic && ns.DynamicBucketTemplate != nil {
		t.Fatalf("Did not configure a dynamic bucket for namespace %v", namespace)
	}

	if len(ns.Buckets) != numBuckets {
		t.Fatal("Expected %v named buckets for namespace %v; found %v", numBuckets, namespace, len(ns.Buckets))
	}
}

func assertBucket(t *testing.T, b *BucketConfig, size, fillRate, waitTimeoutMillis, maxIdleMillis int) {
	if b == nil {
		t.Fatal("Bucket doesn't exist")
	}
	if b.FillRate != fillRate {
		t.Fatalf("Expected fill_rate of %v; was %v", fillRate, b.FillRate)
	}

	if b.WaitTimeoutMillis != waitTimeoutMillis {
		t.Fatalf("Expected wait_timeout_millis of %v; was %v", waitTimeoutMillis, b.WaitTimeoutMillis)
	}

	if b.MaxIdleMillis != maxIdleMillis {
		t.Fatalf("Expected max_idle_millis of %v; was %v", maxIdleMillis, b.MaxIdleMillis)
	}

	if b.Size != size {
		t.Fatalf("Expected bucket size of %v; was %v", size, b.Size)
	}
}

func TestNonexistentFile(t *testing.T) {
	test.ExpectingPanic(t, func() {
		_ = ReadConfigFromFile("/does/not/exist")
	})
}


