package redis

import (
	"testing"
	"reflect"
	"errors"
	"fmt"
	"github.com/maniksurtani/quotaservice/protos/config"
	"github.com/maniksurtani/quotaservice/config"
)

func TestAppendKeys(t *testing.T) {
	keys := make([]string, 0)
	expected := make(map[string]*struct{})
	buckets := []string{"a", "b", "c"}

	for _, b := range buckets {
		keys = appendRedisKeys(keys, "ns", b)

		expected[toRedisKey("ns", b, tokensNextAvblNanosSuffix)] = nil
		expected[toRedisKey("ns", b, accumulatedTokensSuffix)] = nil
	}

	if len(keys) != 6 {
		t.Fatalf("Expected 6 keys, got %v", keys)
	}

	for _, k := range keys {
		if _, exists := expected[k]; !exists {
			t.Fatalf("Expected %v in %v", k, expected)
		}
	}
}

func TestChunking(t *testing.T) {
	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "k"}

	expected := make([][]string, 4)
	expected[0] = []string{"a", "b", "c"}
	expected[1] = []string{"d", "e", "f"}
	expected[2] = []string{"g", "h", "i"}
	expected[3] = []string{"k"}

	invocation := 0

	err := chunkKeys(keys, func(s []string) error {
		if !reflect.DeepEqual(s, expected[invocation]) {
			return errors.New(fmt.Sprintf("Expected %v but was %v", expected[invocation], s))
		}
		invocation++
		return nil
	}, 3)

	if err != nil {
		t.Fatal("Caught error.", err)
	}

	if invocation != 4 {
		t.Fatalf("Expected 4 invocations; only got %v", invocation)
	}
}

func TestFindingKeysToDelete(t *testing.T) {
	existing := make(map[string][]string)
	existing["a"] = []string{"1", "2", "3"}
	existing["b"] = []string{"1", "2", "3"}
	existing["c"] = []string{"1", "2", "3", config.DefaultBucketName}
	existing[config.GlobalNamespace] = []string{config.DefaultBucketName}

	expected := make(map[string]*struct{})

	expected[toRedisKey("a", "3", tokensNextAvblNanosSuffix)] = nil
	expected[toRedisKey("a", "3", accumulatedTokensSuffix)] = nil
	expected[toRedisKey("b", "1", tokensNextAvblNanosSuffix)] = nil
	expected[toRedisKey("b", "1", accumulatedTokensSuffix)] = nil
	expected[toRedisKey("b", "2", tokensNextAvblNanosSuffix)] = nil
	expected[toRedisKey("b", "2", accumulatedTokensSuffix)] = nil
	expected[toRedisKey("b", "3", tokensNextAvblNanosSuffix)] = nil
	expected[toRedisKey("b", "3", accumulatedTokensSuffix)] = nil
	expected[toRedisKey("c", "3", tokensNextAvblNanosSuffix)] = nil
	expected[toRedisKey("c", "3", accumulatedTokensSuffix)] = nil


	cfg := make(map[string]*quotaservice_configs.NamespaceConfig)
	cfg["a"] = config.NewDefaultNamespaceConfig("a")
	cfg["a"].Buckets["1"] = config.NewDefaultBucketConfig("1")
	cfg["a"].Buckets["2"] = config.NewDefaultBucketConfig("2")

	cfg["c"] = config.NewDefaultNamespaceConfig("c")
	cfg["c"].Buckets["1"] = config.NewDefaultBucketConfig("1")
	cfg["c"].Buckets["2"] = config.NewDefaultBucketConfig("2")
	cfg["c"].Buckets["4"] = config.NewDefaultBucketConfig("4")

	cfg["d"] = config.NewDefaultNamespaceConfig("d")
	cfg["d"].Buckets["1"] = config.NewDefaultBucketConfig("1")

	keys := findKeysToDelete(existing, cfg)

	if len(expected) != len(keys) {
		t.Fatal("Expected and results have different element count")
	}

	for _, k := range keys {
		if _, exists := expected[k] ; !exists {
			t.Fatalf("Expected %v in %v", k, expected)
		}
	}
}