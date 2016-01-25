package server
import (
	"testing"
	"time"
)

func TestBucket(t *testing.T) {
	StartTicker()
	b := NewBucket("test", 100)
	b.Start()
	b.Acquire(50)
	time.Sleep(2 * time.Second)
	b.Acquire(25)
	b.Stop()

	b = NewBucket("test2", 20)
	b.Start()
	b.Acquire(50)
	time.Sleep(2 * time.Second)
	b.Acquire(25)

	b2 := NewBucket("test3", 20)
	b2.Start()
	b2.Acquire(50)
	time.Sleep(2 * time.Second)
	b2.Acquire(25)

	b.Stop()
	b2.Stop()

	StopTicker()
}

