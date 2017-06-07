// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

// Package redis implements token buckets backed by Redis, inspired by the algorithms used in Guava's
// RateLimiter library - https://github.com/google/guava/blob/master/guava/src/com/google/common/util/concurrent/RateLimiter.java
package redis

import (
	"strconv"
	"time"

	"gopkg.in/redis.v5"

	"github.com/maniksurtani/quotaservice"
	"github.com/maniksurtani/quotaservice/logging"

	pbconfig "github.com/maniksurtani/quotaservice/protos/config"
)

// redisBucket is an interface that defines the two different bucket types used with Redis: static and dynamic buckets.
type redisBucket interface {
}

// configAttributes represents certain values from a pbconfig.BucketConfig, represented as strings, for easy use as
// parameters to a Redis call.
type configAttributes struct {
	nanosBetweenTokens          string
	maxTokensToAccumulate       string
	maxIdleTimeMillis           string
	maxDebtNanos                string
	*quotaservice.DefaultBucket // Extension for default methods on interface
}

// abstractBucket contains attributes common to both static and dynamic buckets.
type abstractBucket struct {
	*configAttributes
	cfg     *pbconfig.BucketConfig
	factory *bucketFactory
	keys    []string
}

func (a *abstractBucket) Config() *pbconfig.BucketConfig {
	return a.cfg
}

func (a *abstractBucket) Take(requested int64, maxWaitTime time.Duration) (time.Duration, bool) {
	currentTimeNanos := strconv.FormatInt(time.Now().UnixNano(), 10)

	args := []interface{}{currentTimeNanos, a.nanosBetweenTokens, a.maxTokensToAccumulate,
		strconv.FormatInt(requested, 10), strconv.FormatInt(maxWaitTime.Nanoseconds(), 10),
		a.maxIdleTimeMillis, a.maxDebtNanos}

	keepTrying := true
	var waitTime time.Duration
	var err error
	for attempt := 0; keepTrying && attempt < a.factory.connectionRetries; attempt++ {
		client := a.factory.Client().(*redis.Client)
		res := client.EvalSha(a.factory.scriptSHA, a.keys, args...)
		switch waitTimeNanos := res.Val().(type) {
		case int64:
			waitTime = time.Nanosecond * time.Duration(waitTimeNanos)
			keepTrying = false
		default:
			err = res.Err()
			if unknownCloseError(err) {
				logging.Printf("Unknown response '%v' of type %T. Full result %+v",
					waitTimeNanos, waitTimeNanos, res)
			}

			a.factory.reconnectToRedis(client)
		}
	}

	if keepTrying {
		// TODO(kaneda): Gracefully handle Redis access errors?
		logging.Fatalf("Couldn't reconnect to Redis, even after %v attempts with error %+v",
			a.factory.connectionRetries, err)
	}

	if waitTime < 0 {
		// Timed out
		return 0, false
	}

	return waitTime, true
}

// staticBucket is an implementation of a redisBucket for use with static, named buckets.
type staticBucket struct {
	*abstractBucket
}

func (s *staticBucket) Dynamic() bool {
	return false
}

// dynamicBucket is an implementation of a redisBucket for use with dynamic buckets created from a template.
type dynamicBucket struct {
	*abstractBucket
}

func (d *dynamicBucket) Dynamic() bool {
	return true
}

func (d *dynamicBucket) Destroy() {
	// decrease ref-count common
	d.factory.Lock()
	defer d.factory.Unlock()
	d.factory.refcounts[d.cfg.Namespace]--

	if d.factory.refcounts[d.cfg.Namespace] < 0 {
		logging.Fatalf("Ref counts for %v went negative! refcounts=%+v sharedAttributes=%+v", d.cfg.Namespace, d.factory.refcounts, d.factory.sharedAttributes)
	}

	// If ref-count hits 0, remove common bucket fields
	if d.factory.refcounts[d.cfg.Namespace] == 0 {
		delete(d.factory.sharedAttributes, d.cfg.Namespace)
		delete(d.factory.refcounts, d.cfg.Namespace)
	}
}
