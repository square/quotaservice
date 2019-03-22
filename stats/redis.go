// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package stats

import (
	"fmt"
	"time"

	"gopkg.in/redis.v5"

	"github.com/square/quotaservice/events"
	"github.com/square/quotaservice/logging"
)

type redisListener struct {
	client *redis.Client
}

// NewRedisStatsListener creates a redis-backed stats
// listener with the passed in redis.Options.
func NewRedisStatsListener(redisOpts *redis.Options) Listener {
	client := redis.NewClient(redisOpts)
	_, err := client.Ping().Result()

	if err != nil {
		logging.Fatalf("RedisStatsListener: cannot connect to Redis, %v", err)
	}

	return &redisListener{client}
}

func (l *redisListener) redisTopList(key string) []*BucketScore {
	results, err := l.client.ZRevRangeWithScores(key, 0, 10).Result()

	if err != nil && err.Error() != "redis: nil" {
		logging.Printf("RedisStatsListener.TopList error (%s) %v", key, err)
		return emptyArr
	}

	arr := make([]*BucketScore, len(results))

	for i, item := range results {
		arr[i] = &BucketScore{item.Member.(string), int64(item.Score)}
	}

	return arr
}

func statsNamespace(key, namespace string) string {
	return fmt.Sprintf("stats:%s:%s", namespace, key)
}

// TopHits is implemented for stats.Listener
// TopHits returns a sorted list of the 10 buckets with the highest # of hits
// in the specified namespace within the current bucketed hour
func (l *redisListener) TopHits(namespace string) []*BucketScore {
	return l.redisTopList(statsNamespace("hits", namespace))
}

// TopMisses is implemented for stats.Listener
// TopMisses returns a sorted list of the 10 buckets with the highest # of misses
// in the specified namespace within the current bucketed hour
func (l *redisListener) TopMisses(namespace string) []*BucketScore {
	return l.redisTopList(statsNamespace("misses", namespace))
}

// Get is implemented for stats.Listener
// Get returns the hits and misses for a bucket in the specified namespace
// within the current bucketed hour
func (l *redisListener) Get(namespace, bucket string) *BucketScores {
	scores := &BucketScores{0, 0}

	value, err := l.client.ZScore(statsNamespace("misses", namespace), bucket).Result()

	if err != nil && err.Error() != "redis: nil" {
		logging.Printf("RedisStatsListener.Get error (%s, %s) %v", namespace, bucket, err)
	} else {
		scores.Misses = int64(value)
	}

	value, err = l.client.ZScore(statsNamespace("hits", namespace), bucket).Result()

	if err != nil && err.Error() != "redis: nil" {
		logging.Printf("RedisStatsListener.Get error (%s, %s) %v", namespace, bucket, err)
	} else {
		scores.Hits = int64(value)
	}

	return scores
}

func nearestHour() time.Time {
	return time.Now().Add(time.Hour).Truncate(time.Hour)
}

// HandleEvent is implemented for stats.Listener
// HandleEvent consumes dynamic bucket events (see events.Event)
func (l *redisListener) HandleEvent(event events.Event) {
	if !event.Dynamic() {
		return
	}

	var key string
	var numTokens int64 = 1

	switch event.EventType() {
	case events.EVENT_BUCKET_MISS:
		key = "misses"
	case events.EVENT_TOKENS_SERVED:
		numTokens = event.NumTokens()
		key = "hits"
	default:
		return
	}

	namespace := statsNamespace(key, event.Namespace())
	bucket := event.BucketName()

	var incr *redis.FloatCmd
	_, err := l.client.Pipelined(func(pipe *redis.Pipeline) error {
		incr = pipe.ZIncrBy(namespace, float64(numTokens), bucket)
		pipe.ExpireAt(namespace, nearestHour())
		return nil
	})

	if err != nil || incr.Err() != nil {
		logging.Printf("RedisStatsListener.HandleEvent error (%s, %s, %d) %v, %v",
			namespace, bucket, numTokens, err, incr.Err())
	}
}
