// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package config

import (
	"testing"
	"reflect"
	"fmt"
	"github.com/maniksurtani/quotaservice/test/helpers"
)

const cfgYaml = `namespaces:
  no_default_no_dynamic:
    buckets:
      one:
        fill_rate: 321
        wait_timeout_millis: 9999
        max_idle_millis: 20000
        max_debt_millis: 30000
      with_defaults:
        size: 100
  only_dynamic:
    max_dynamic_buckets: 50
    dynamic_bucket_template:
      fill_rate: 999
      wait_timeout_millis: 8888
      max_idle_millis: 30000
      max_tokens_per_request: 5
  only_default:
    default_bucket:
      fill_rate: 800
      wait_timeout_millis: 7777
      max_idle_millis: 40000
`

func TestConfig(t *testing.T) {
	cfg := readConfigFromBytes([]byte(cfgYaml))

	if cfg.GlobalDefaultBucket != nil {
		t.Fatal("Did not configure a global default bucket")
	}

	if len(cfg.Namespaces) != 3 {
		t.Fatalf("Expected 3 namespaces; was %v", len(cfg.Namespaces))
	}

	namespace := "no_default_no_dynamic"
	ns := cfg.Namespaces[namespace]

	assertNamespace(t, namespace, ns, 2, false, false, 0)
	assertBucket(t, ns.Buckets["one"], 100, 321, 9999, 20000, 30000, 321)
	assertBucket(t, ns.Buckets["with_defaults"], 100, 50, 1000, -1, 10000, 50)

	namespace = "only_dynamic"
	ns = cfg.Namespaces[namespace]

	assertNamespace(t, namespace, ns, 0, false, true, 50)
	assertBucket(t, ns.DynamicBucketTemplate, 100, 999, 8888, 30000, 10000, 5)

	namespace = "only_default"
	ns = cfg.Namespaces[namespace]

	assertNamespace(t, namespace, ns, 0, true, false, 0)
	assertBucket(t, ns.DefaultBucket, 100, 800, 7777, 40000, 10000, 800)
}

func assertNamespace(t *testing.T, namespace string, ns *NamespaceConfig, numBuckets int, expectDefault, expectDynamic bool, maxDynamic int) {
	if !expectDefault && ns.DefaultBucket != nil {
		t.Fatalf("Did not configure a default bucket for namespace %v", namespace)
	}

	if !expectDynamic && ns.DynamicBucketTemplate != nil {
		t.Fatalf("Did not configure a dynamic bucket for namespace %v", namespace)
	}

	if len(ns.Buckets) != numBuckets {
		t.Fatalf("Expected %v named buckets for namespace %v; found %v", numBuckets, namespace, len(ns.Buckets))
	}

	if ns.MaxDynamicBuckets != maxDynamic {
		t.Fatalf("Expected %v max_dynamic_buckets for namespace %v; found %v", maxDynamic, namespace, ns.MaxDynamicBuckets)
	}
}

func assertBucket(t *testing.T, b *BucketConfig, size, fillRate, waitTimeoutMillis, maxIdleMillis, maxDebtMillis, maxTokensPerRequest int64) {
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

	if b.MaxDebtMillis != maxDebtMillis {
		t.Fatalf("Expected max_debt_millis of %v; was %v", maxDebtMillis, b.MaxDebtMillis)
	}

	if b.Size != size {
		t.Fatalf("Expected bucket size of %v; was %v", size, b.Size)
	}

	if b.MaxTokensPerRequest != maxTokensPerRequest {
		t.Fatalf("Expected max tokens per request of %v; was %v", maxTokensPerRequest, b.MaxTokensPerRequest)
	}
}

func TestNonexistentFile(t *testing.T) {
	helpers.ExpectingPanic(t, func() {
		_ = ReadConfigFromFile("/does/not/exist")
	})
}

func TestToAndFromProtos(t *testing.T) {
	osc := readConfigFromBytes([]byte(cfgYaml))
	p := osc.ToProto()
	recreated := FromProto(p)

	if osc.Version != recreated.Version {
		t.Fatal("version field mismatch")
	}

	testBuckets(t, DefaultBucketName, osc.GlobalDefaultBucket, recreated.GlobalDefaultBucket)

	if len(osc.Namespaces) != len(recreated.Namespaces) {
		t.Fatalf("Different number of namespaces. Original %v, recreated %v", len(osc.Namespaces), len(recreated.Namespaces))
	}

	for n, n1 := range osc.Namespaces {
		n2 := recreated.Namespaces[n]
		if n2 == nil {
			t.Fatalf("Namespace %v doesn't exist on replica", n)
		}

		testNamespaces(t, n, n1, n2)
	}
}

func testBuckets(t *testing.T, name string, b1, b2 *BucketConfig) {
	fmt.Printf("    Testing bucket %v\n", name)

	if bothNils(t, b1, b2) {
		return
	}

	if b1.FillRate != b2.FillRate {
		t.Fatal("Fill rate mismatch")
	}
	if b1.Size != b2.Size {
		t.Fatal("Size mismatch")
	}
	if b1.MaxDebtMillis != b2.MaxDebtMillis {
		t.Fatal("MaxDebtMillis mismatch")
	}
	if b1.MaxIdleMillis != b2.MaxIdleMillis {
		t.Fatal("MaxIdleMillis mismatch")
	}
	if b1.MaxTokensPerRequest != b2.MaxTokensPerRequest {
		t.Fatal("MaxTokensPerRequest mismatch")
	}
	if b1.WaitTimeoutMillis != b2.WaitTimeoutMillis {
		t.Fatal("WaitTimeoutMillis mismatch")
	}
}

func testNamespaces(t *testing.T, name string, n1, n2 *NamespaceConfig) {
	fmt.Printf("  Testing namespace %v\n", name)

	if bothNils(t, n1, n2) {
		return
	}

	if n1.MaxDynamicBuckets != n2.MaxDynamicBuckets {
		t.Fatal("MaxDynamicBuckets mismatch")
	}
	testBuckets(t, DefaultBucketName, n1.DefaultBucket, n2.DefaultBucket)
	testBuckets(t, DynamicBucketTemplateName, n1.DynamicBucketTemplate, n2.DynamicBucketTemplate)

	if len(n1.Buckets) != len(n2.Buckets) {
		t.Fatal("Different number of buckets")
	}

	for n, b1 := range n1.Buckets {
		b2 := n2.Buckets[n]
		if b2 == nil {
			t.Fatalf("Bucket %v doesn't exist on namespace", n)
		}

		testBuckets(t, n, b1, b2)
	}
}

func bothNils(t *testing.T, o1, o2 interface{}) bool {
	if (o1 == nil && o2 == nil) || (reflect.ValueOf(o1).IsNil() && reflect.ValueOf(o2).IsNil()) {
		return true
	}
	if (o1 == nil || reflect.ValueOf(o1).IsNil()) || (o2 == nil || reflect.ValueOf(o2).IsNil()) {
		t.Fatalf("o1 = %+v ; o2 = %+v", o1, o2)
	}
	return false
}
