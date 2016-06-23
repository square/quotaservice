// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package config

import (
	"github.com/maniksurtani/quotaservice/test/helpers"
	"testing"

	pbconfig "github.com/maniksurtani/quotaservice/protos/config"
	"strings"
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
	cfg := ReadConfig(strings.NewReader(cfgYaml))

	if cfg.GlobalDefaultBucket != nil {
		t.Fatal("Did not configure a global default bucket")
	}

	if len(cfg.Namespaces) != 3 {
		t.Fatalf("Expected 3 namespaces; was %v", len(cfg.Namespaces))
	}

	namespace := "no_default_no_dynamic"
	ns := cfg.Namespaces[namespace]

	assertNamespace(t, namespace, ns, 2, false, false, 0)
	assertBucket(t, "one", namespace, ns.Buckets["one"], 100, 321, 9999, 20000, 30000, 321)
	assertBucket(t, "with_defaults", namespace, ns.Buckets["with_defaults"], 100, 50, 1000, -1, 10000, 50)

	namespace = "only_dynamic"
	ns = cfg.Namespaces[namespace]

	assertNamespace(t, namespace, ns, 0, false, true, 50)
	assertBucket(t, DynamicBucketTemplateName, namespace, ns.DynamicBucketTemplate, 100, 999, 8888, 30000, 10000, 5)

	namespace = "only_default"
	ns = cfg.Namespaces[namespace]

	assertNamespace(t, namespace, ns, 0, true, false, 0)
	assertBucket(t, DefaultBucketName, namespace, ns.DefaultBucket, 100, 800, 7777, 40000, 10000, 800)
}

func assertNamespace(t *testing.T, namespace string, ns *pbconfig.NamespaceConfig, numBuckets int, expectDefault, expectDynamic bool, maxDynamic int32) {
	if namespace != ns.Name {
		t.Fatalf("namespace.Name is %v and not %v", ns.Name, namespace)
	}

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

func assertBucket(t *testing.T, name, namespace string, b *pbconfig.BucketConfig, size, fillRate, waitTimeoutMillis, maxIdleMillis, maxDebtMillis, maxTokensPerRequest int64) {
	if b == nil {
		t.Fatal("Bucket doesn't exist")
	}

	if name != b.Name {
		t.Fatalf("bucket.Name is %v and not %v", b.Name, name)
	}

	if namespace != b.Namespace {
		t.Fatalf("bucket.Namespace is %v and not %v", b.Namespace, namespace)
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
