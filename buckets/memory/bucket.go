// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

// Package memory implements token buckets in memory, inspired by the algorithms used in Guava's
// RateLimiter library - https://github.com/google/guava/blob/master/guava/src/com/google/common/util/concurrent/RateLimiter.java
package memory

import (
	"time"

	"github.com/maniksurtani/quotaservice/buckets"
	"github.com/maniksurtani/quotaservice/configs"
	"github.com/maniksurtani/quotaservice/logging"
)

type bucketFactory struct {
	cfg *configs.ServiceConfig
}

func (bf *bucketFactory) Init(cfg *configs.ServiceConfig) {
	bf.cfg = cfg
}

func (bf *bucketFactory) NewBucket(namespace, bucketName string, cfg *configs.BucketConfig, dyn bool) buckets.Bucket {
	// fill rate is tokens-per-second.
	bucket := &tokenBucket{
		ActivityChannel: buckets.NewActivityChannel(),
		dynamic: dyn,
		cfg: cfg,
		nanosBetweenTokens: 1e9 / cfg.FillRate,
		accumulatedTokens: cfg.Size, // Start full
		fullName: buckets.FullyQualifiedName(namespace, bucketName),
		waitTimer: make(chan *waitTimeReq),
		closer: make(chan struct{})}

	go bucket.waitTimeLoop()

	return bucket
}

func NewBucketFactory() buckets.BucketFactory {
	return &bucketFactory{}
}

// tokenBucket is a single-threaded implementation. A single goroutine updates the values of
// tokensNextAvailable and accumulatedTokens. When requesting tokens, Take() puts a request on
// the waitTimer channel, and listens on the response channel in the request for a result. The
// goroutine is shut down when Destroy() is called on this bucket. In-flight requests will be
// served, but new requests will not.
type tokenBucket struct {
	buckets.ActivityChannel
	dynamic           bool
	cfg               *configs.BucketConfig
	nanosBetweenTokens,
	tokensNextAvailableNanos,
	accumulatedTokens int64
	fullName          string
	waitTimer         chan *waitTimeReq
	closer            chan struct{}
}

// waitTimeReq is a request that you put on the channel for the waitTimer goroutine to pick up and
// process.
type waitTimeReq struct {
	requested, maxWaitTimeNanos int64
	response                    chan int64
}

func (b *tokenBucket) Take(numTokens int64, maxWaitTime time.Duration) (waitTime time.Duration) {
	rsp := make(chan int64, 1)
	b.waitTimer <- &waitTimeReq{numTokens, maxWaitTime.Nanoseconds(), rsp}
	waitTimeNanos := <-rsp

	waitTime = time.Duration(waitTimeNanos) * time.Nanosecond
	if waitTime > maxWaitTime && maxWaitTime > 0 {
		waitTime = -1
	}

	return
}

func (b *tokenBucket) calcWaitTime(requested, maxWaitTimeNanos int64) (waitTimeNanos int64) {
	currentTimeNanos := time.Now().UnixNano()
	tna := b.tokensNextAvailableNanos
	ac := b.accumulatedTokens

	var freshTokens int64 = 0

	if currentTimeNanos > tna {
		freshTokens = (currentTimeNanos - tna) / b.nanosBetweenTokens
		ac = min(b.cfg.Size, ac + freshTokens)
		tna = currentTimeNanos
	}

	waitTimeNanos = tna - currentTimeNanos
	accumulatedTokensUsed := min(ac, requested)
	tokensToWaitFor := requested - accumulatedTokensUsed
	futureWaitNanos := tokensToWaitFor * b.nanosBetweenTokens

	tna += futureWaitNanos
	ac -= accumulatedTokensUsed

	if (tna - currentTimeNanos > b.cfg.MaxDebtMillis * 1e6) || (waitTimeNanos > 0 && waitTimeNanos > maxWaitTimeNanos && maxWaitTimeNanos > 0) {
		waitTimeNanos = -1
	} else {
		b.tokensNextAvailableNanos = tna
		b.accumulatedTokens = ac
	}

	return waitTimeNanos
}

func min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func (b *tokenBucket) waitTimeLoop() {
	keepRunning := true
	for ; keepRunning; {
		select {
		case req := <-b.waitTimer:
			w := b.calcWaitTime(req.requested, req.maxWaitTimeNanos)
			req.response <- w
		case <-b.closer:
			keepRunning = false
			logging.Printf("Garbage collecting bucket %v", b.fullName)
		}
	}
}

func (b *tokenBucket) Config() *configs.BucketConfig {
	return b.cfg
}

func (b *tokenBucket) Dynamic() bool {
	return b.dynamic
}

func (b *tokenBucket) Destroy() {
	// Signal the waitTimeLoop to exit
	b.closer <- struct{}{}
}
