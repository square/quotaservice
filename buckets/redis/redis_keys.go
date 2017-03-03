package redis

import (
	"fmt"
	"github.com/maniksurtani/quotaservice/config"
	"github.com/maniksurtani/quotaservice/logging"
	qsc "github.com/maniksurtani/quotaservice/protos/config"
	"gopkg.in/redis.v3"
	"math"
	"strings"
)

const (
	REDIS_DEL_CHUNK_SIZE = 10000
	REDIS_SCAN_PAGE_SIZE = 10000
)

// existingBuckets scans Redis and returns a mapping of namespaces to a slice of bucket names in that namespace, as
// observed in Redis.
func existingBuckets(r *redis.Client) (map[string][]string, error) {
	buckets := make(map[string][]string)

	var next int64
	var err error

	for {
		var keys []string
		next, keys, err = r.Scan(next, fmt.Sprintf("*:%v", accumulatedTokensSuffix), REDIS_SCAN_PAGE_SIZE).Result()
		if err != nil {
			return nil, err
		}

		// Iterate over these keys, extract namespace and bucket
		for _, k := range keys {
			addBucket(k, buckets)
		}

		if next == 0 {
			// We're done with the scan.
			break
		}
	}

	return buckets, nil
}

// addBucket parses a passed-in Redis key to extract the key's namespace and bucket name, adding it to the map passed
// in.
func addBucket(key string, buckets map[string][]string) {
	if !strings.Contains(key, ":") {
		return
	}

	parts := strings.Split(key, ":")
	ns, b := parts[0], parts[1]

	_, ok := buckets[ns]
	if !ok {
		buckets[ns] = make([]string, 0)
	}

	buckets[ns] = append(buckets[ns], b)
}

// deleteUnknown deletes all keys related to unknown buckets in Redis. `existing` is a collection of namespaces and
// bucket names that are known to exist in Redis, based on the results of calling `existingBuckets()`, and `cfg` is the
// map of NamespaceConfigs containing currently configured namespaces and buckets.
func deleteUnknown(r *redis.Client, existing map[string][]string, cfg map[string]*qsc.NamespaceConfig) error {
	redisKeysToDelete := findKeysToDelete(existing, cfg)

	return chunkKeys(redisKeysToDelete, func(keys []string) error {

		numDel, err := r.Del(keys...).Result()
		if err != nil {
			return err
		}

		logging.Printf("Purged %v unused keys from Redis", numDel)
		return nil
	}, REDIS_DEL_CHUNK_SIZE)
}

// findKeysToDelete inspects a map of namespace to bucket names (the starting collection of buckets), and a
// NamespaceConfig. All buckets in the former, that do not exist in the latter, are converted to their Redis keys and
// added to the slice that is returned. Note that buckets that aren't statically defined in a namespace, but in a
// namespace that allows dynamic buckets, aren't considered candidates for removal. Global/Default buckets are never
// considered for removal.
func findKeysToDelete(existing map[string][]string, cfg map[string]*qsc.NamespaceConfig) []string {
	redisKeysToDelete := make([]string, 0)

	for nsName, bucketNames := range existing {
		if nsName == config.GlobalNamespace {
			continue
		}

		ns, exists := cfg[nsName]
		for _, bucketName := range bucketNames {
			if !exists {
				// The entire namespace doesn't exist.
				redisKeysToDelete = appendRedisKeys(redisKeysToDelete, nsName, bucketName)
			} else if ns.DynamicBucketTemplate == nil {
				// No dynamic buckets in this namespace. Check all buckets.
				_, bucketExists := ns.Buckets[bucketName]
				if !bucketExists && bucketName != config.DefaultBucketName {
					// The bucket doesn't exist.
					redisKeysToDelete = appendRedisKeys(redisKeysToDelete, nsName, bucketName)
				}
			}
		}
	}

	return redisKeysToDelete
}

// appendRedisKeys creates all keys we expect to see for a given namespace/bucket name combo, and adds these keys to a
// slice of keys passed in. Similar to Go's builtin `append()` function, you should always assign the return value of
// this function to the slice variable you intend to use.
func appendRedisKeys(keys []string, nsName, bName string) []string {
	keys = append(keys, toRedisKey(nsName, bName, tokensNextAvblNanosSuffix))
	return append(keys, toRedisKey(nsName, bName, accumulatedTokensSuffix))
}

// chunkKeys breaks up a slice of keys passed in into a smaller chunks, according to `chunkSize`, and invokes `callback`
// with each chunk as a parameter. If the callback function returns an error, this function exits immediately,
// propagating that error.
func chunkKeys(keys []string, callback func([]string) error, chunkSize int) error {
	// Chunk the keys to be deleted.
	numKeys := len(keys)
	loops := int(math.Ceil(float64(numKeys) / float64(chunkSize)))

	for i := 0; i < loops; i++ {
		chunkStart := i * chunkSize
		chunkEnd := (i + 1) * chunkSize

		if chunkEnd > numKeys {
			chunkEnd = numKeys
		}

		err := callback(keys[chunkStart:chunkEnd])

		if err != nil {
			return err
		}
	}

	return nil
}
