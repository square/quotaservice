// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package stats

import (
	"sort"

	"github.com/maniksurtani/quotaservice/events"
)

type namespaceStats struct {
	hits, misses map[string]*BucketScore
}

type memoryListener struct {
	namespaces map[string]*namespaceStats
}

func NewMemoryStatsListener() *memoryListener {
	return &memoryListener{make(map[string]*namespaceStats)}
}

func (l *memoryListener) bucketScoreTop10(scoreMap map[string]*BucketScore) []*BucketScore {
	arr := make(BucketScoreArray, 0)

	for _, value := range scoreMap {
		arr = append(arr, value)
	}

	sort.Sort(arr)
	length := len(arr)

	if length > 10 {
		length = 10
	}

	return arr[0:length]
}

// TopHits is implemented for stats.Listener
// TopHits returns a sorted list of the 10 buckets with the highest # of hits
// in the specified namespace
func (l *memoryListener) TopHits(namespace string) []*BucketScore {
	stats, ok := l.namespaces[namespace]

	if !ok {
		return emptyArr
	}

	return l.bucketScoreTop10(stats.hits)
}

// TopMisses is implemented for stats.Listener
// TopMisses returns a sorted list of the 10 buckets with the highest # of misses
// in the specified namespace
func (l *memoryListener) TopMisses(namespace string) []*BucketScore {
	stats, ok := l.namespaces[namespace]

	if !ok {
		return emptyArr
	}

	return l.bucketScoreTop10(stats.misses)
}

// Get is implemented for stats.Listener
// Get returns the hits and misses for a bucket in the specified namespace
func (l *memoryListener) Get(namespace, bucket string) *BucketScores {
	stats, ok := l.namespaces[namespace]

	if !ok {
		return emptyBucketScores
	}

	scores := &BucketScores{0, 0}

	if hitValue, ok := stats.hits[bucket]; ok {
		scores.Hits = hitValue.Score
	}

	if missValue, ok := stats.misses[bucket]; ok {
		scores.Misses = missValue.Score
	}

	return scores
}

// HandleEvent is implemented for stats.Listener
// HandleEvent consumes dynamic bucket events (see events.Event)
func (l *memoryListener) HandleEvent(event events.Event) {
	if !event.Dynamic() {
		return
	}

	namespace := event.Namespace()

	if _, ok := l.namespaces[namespace]; !ok {
		l.namespaces[namespace] = &namespaceStats{
			make(map[string]*BucketScore),
			make(map[string]*BucketScore)}
	}

	stats := l.namespaces[namespace]

	var statsBucket map[string]*BucketScore
	var numTokens int64 = 1

	switch event.EventType() {
	case events.EVENT_BUCKET_MISS:
		statsBucket = stats.misses
	case events.EVENT_TOKENS_SERVED:
		numTokens = event.NumTokens()
		statsBucket = stats.hits
	default:
		return
	}

	key := event.BucketName()

	if _, ok := statsBucket[key]; !ok {
		statsBucket[key] = &BucketScore{key, 0}
	}

	statsBucket[key].Score += numTokens
}
