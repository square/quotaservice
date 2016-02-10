// tokenbucket.go (c) 2014 David Rook - all rights reserved
//
// see README-tokenbucket-pkg.md
//  also http://en.wikipedia.org/wiki/Token_bucket
package tokenbucket

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type TokenBucket struct {
	fillInterval  time.Duration
	capacity      float64
	lastCount     float64
	lastCheckTime time.Time
	lock          sync.Mutex
}

// New() creates a tokenbucket object
func New(fillTime time.Duration, capacity float64) *TokenBucket {
	if fillTime.Nanoseconds() <= int64(0) {
		log.Fatalf("arguments to tokenbucket.New() must be positive non-zero\n")
	}
	if capacity <= 0 {
		log.Fatalf("arguments to tokenbucket.New() must be positive non-zero\n")
	}
	var tb TokenBucket
	tb.fillInterval = fillTime
	tb.capacity = capacity
	tb.lastCount = capacity
	tb.lastCheckTime = time.Now()
	return &tb
}

func (t *TokenBucket) dump() {
	t.lock.Lock()
	fmt.Printf("Interval[%v] Cap[%g] lastCount[%g] lastCheckTime[%v]\n",
		t.fillInterval, t.capacity, t.lastCount, t.lastCheckTime)
	t.lock.Unlock()
}

// FillRate() returns number of tokens per second
func (t *TokenBucket) FillRate() float64 {
	return float64(time.Second.Nanoseconds()) / float64(t.fillInterval.Nanoseconds())
}

// Take() returns the time to wait before tokens are available.
//  Calling Take commits to take them, tokens can't be put back.
func (t *TokenBucket) Take(icount int64) time.Duration {
	count := float64(icount)
	now := time.Now()
	t.lock.Lock()
	t.lastCount += float64(now.Sub(t.lastCheckTime).Nanoseconds()) / float64(t.fillInterval.Nanoseconds())
	t.lastCheckTime = now
	if t.lastCount > t.capacity {
		t.lastCount = t.capacity
	}
	t.lastCount -= count
	var delay time.Duration
	if t.lastCount <= 0 {
		delay = time.Duration(-t.lastCount * float64(t.fillInterval.Nanoseconds()))
		//fmt.Printf("Take Delay = %v\n",delay)
	}
	t.lock.Unlock()
	return delay
}
