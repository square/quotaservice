// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

// Package redis implements token buckets backed by Redis, inspired by the algorithms used in Guava's
// RateLimiter library - https://github.com/google/guava/blob/master/guava/src/com/google/common/util/concurrent/RateLimiter.java
package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/square/quotaservice"
	"github.com/square/quotaservice/logging"
	pbconfig "github.com/square/quotaservice/protos/config"
)

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

func (a *abstractBucket) Take(ctx context.Context, requested int64, maxWaitTime time.Duration) (time.Duration, bool, error) {
	currentTimeNanos := strconv.FormatInt(time.Now().UnixNano(), 10)

	maxIdleTimeMillis := a.maxIdleTimeMillis
	if a.maxIdleTimeMillis == "0" {
		// bucket MaxIdleMillis was not set; fall back to factory setting
		maxIdleTimeMillis = strconv.FormatInt(int64(a.factory.keyMaxIdleTime/time.Millisecond), 10)
	}
	args := []interface{}{currentTimeNanos, a.nanosBetweenTokens, a.maxTokensToAccumulate,
		strconv.FormatInt(requested, 10), strconv.FormatInt(maxWaitTime.Nanoseconds(), 10),
		maxIdleTimeMillis, a.maxDebtNanos}

	client := a.factory.Client().(*redis.Client)
	res := a.takeFromRedis(ctx, client, args)
	if err := res.Err(); err != nil {
		if isRedisClientClosedError(err) {
			logging.Print("Failed to take token from redis because the client was closed, reconnecting")
			a.factory.handleConnectionFailure(client)
		}
		return 0, false, errors.Wrap(err, "failed to take token from redis bucket")
	}

	var waitTime time.Duration
	switch val := res.Val().(type) {
	case int64:
		waitTime = time.Nanosecond * time.Duration(val)
	default:
		return 0, false, errors.Errorf("unknown response of type %[1]T: %[1]v", val)
	}

	if waitTime < 0 {
		// Timed out
		return 0, false, nil
	}

	return waitTime, true, nil
}

func (a *abstractBucket) takeFromRedis(ctx context.Context, client *redis.Client, args []interface{}) *redis.Cmd {
	span, ctx := opentracing.StartSpanFromContext(ctx, "script.Run")
	defer span.Finish()
	return a.factory.script.Run(client, a.keys, args...)
}

var _ quotaservice.Bucket = (*staticBucket)(nil)

// staticBucket is an implementation of quotaservice.Bucket for use with static, named buckets.
type staticBucket struct {
	*abstractBucket
}

func (s *staticBucket) Dynamic() bool {
	return false
}

var _ quotaservice.Bucket = (*dynamicBucket)(nil)

// dynamicBucket is an implementation of quotaservice.Bucket for use with dynamic buckets created from a template.
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
