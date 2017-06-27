// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package stats

import (
	"fmt"

	"github.com/square/quotaservice/events"
)

// Listener is an interface for consuming
// and retrieving dynamic bucket hits and misses
type Listener interface {
	TopHits(string) []*BucketScore
	TopMisses(string) []*BucketScore
	Get(string, string) *BucketScores
	HandleEvent(events.Event)
}

// BucketScores stores a specific bucket's
// stats on hits and misses
type BucketScores struct {
	Hits   int64 `json:"hits"`
	Misses int64 `json:"misses"`
}

// BucketScore stores a specific bucket's
// stats. Used for top-lists.
type BucketScore struct {
	Bucket string `json:"bucket"`
	Score  int64  `json:"value"`
}

var emptyArr []*BucketScore
var emptyBucketScores *BucketScores

func init() {
	emptyArr = make([]*BucketScore, 0)
	emptyBucketScores = &BucketScores{0, 0}
}

func (b *BucketScore) String() string {
	return fmt.Sprintf("{%s, %d}", b.Bucket, b.Score)
}

// BucketScoreArray mplements a sortable BucketScore array
type BucketScoreArray []*BucketScore

func (b BucketScoreArray) Len() int {
	return len(b)
}

func (b BucketScoreArray) Less(i, j int) bool {
	return b[i].Score > b[j].Score
}

func (b BucketScoreArray) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
