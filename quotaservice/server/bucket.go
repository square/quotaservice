package server
import (
	"time"
	"math"
	"log"
)

type Bucket struct {
	Name     string
	fillRate int
	tokens   int
	capacity int
}

// For the ticker
var (
	BucketRegistry map[string]*Bucket = make(map[string]*Bucket) // TODO(manik) default sizes?
	currentStatus status
	stopSignal    chan bool
	ticker        *time.Ticker
)


func NewBucket(name string, fillRate int) *Bucket {
	return &Bucket{Name: name, fillRate: fillRate}
}

func (this *Bucket) Start() {
	this.tokens = this.fillRate
	this.capacity = this.fillRate * 5 // hardcoded for now
	BucketRegistry[this.Name] = this
	log.Printf("Starting bucket '%v'", this.Name)
}

func (this *Bucket) Stop() {
	delete(BucketRegistry, this.Name)
	log.Printf("Stopping bucket '%v'", this.Name)
	this.tokens = 0
}

func (this *Bucket) Acquire(numTokens int) int {
	if this.tokens > numTokens {
		this.tokens -= numTokens
		return numTokens
	} else {
		acquired := this.tokens
		this.tokens = 0
		return acquired
	}
}

func StartTicker() {
	stopSignal = make(chan bool, 1)
	ticker = time.NewTicker(1 * time.Second)
	currentStatus = started
	go fillBuckets()
}

func StopTicker() {
	for n, _ := range BucketRegistry {
		delete(BucketRegistry, n)
	}

	currentStatus = stopped
	stopSignal <- true
	ticker.Stop()
}

func fillBuckets() {
	run := true

	for run {
		select {
		case signal := <-stopSignal:
			log.Printf("Received stop signal: %v", signal)
			run = !signal
		default:
		// Wait for a tick
			<-ticker.C
			for _, bucket := range BucketRegistry {
				fillBucket(bucket)
			}
		}
	}
}

func fillBucket(bucket *Bucket) {
	oldTokens := bucket.tokens
	if bucket.capacity > bucket.tokens {
		bucket.tokens = int(math.Min(float64(bucket.capacity), float64(bucket.fillRate + bucket.tokens)))
		log.Printf("%v received tick; tokens upped from %v to %v.", bucket.Name, oldTokens, bucket.tokens)
	} else {
		log.Printf("%v received tick; tokens at capacity.", bucket.Name)
	}
}


