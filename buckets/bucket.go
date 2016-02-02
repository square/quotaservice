package buckets
import (
	"time"
	"github.com/hotei/tokenbucket"
	"github.com/maniksurtani/quotaservice/logging"
	"github.com/maniksurtani/quotaservice/configs"
)

type TokenBucketsContainer struct {
	Buckets map[string]*tokenBucket
}

type tokenBucket struct {
	tokenbucket.TokenBucket // Embed actual token bucket
	Cfg *configs.BucketConfig
}

var tokenBuckets *TokenBucketsContainer

func InitBuckets(cfg *configs.Configs) *TokenBucketsContainer {
	tokenBuckets = &TokenBucketsContainer{Buckets:make(map[string]*tokenBucket)}
	logging.Print("Initializing buckets")
	for n, b := range cfg.Buckets {
		tokenBuckets.addBucket(n, b)
	}
	logging.Print("Finished initializing buckets")
	return tokenBuckets
}

func (tbc *TokenBucketsContainer) addBucket(name string, cfg *configs.BucketConfig) {
	dur := time.Nanosecond * time.Duration(1000000000 / cfg.FillRate)
	logging.Printf("Creating bucket for name %v with fill duration %v and capacity %v", name, dur, cfg.Size)
	tb := tokenbucket.New(dur, float64(cfg.Size))
	tbc.Buckets[name] = &tokenBucket{
		Cfg: cfg,
		TokenBucket: *tb}
}

func (tb *TokenBucketsContainer) FindBucket(name string) *tokenBucket {
	b := tb.Buckets[name]
	// TODO perform an actual search
	return b
}

//func startFiller() {
//	stopSignal = make(chan bool, 1)
//	ticker = time.NewTicker(1 * time.Second)
//	currentStatus = started
//	go fillBuckets()
//}
//
//func stopFiller() {
//	for n, _ := range BucketRegistry {
//		delete(BucketRegistry, n)
//	}
//
//	currentStatus = stopped
//	stopSignal <- true
//	ticker.Stop()
//}
//
//func fillBuckets() {
//	run := true
//
//	for run {
//		select {
//		case signal := <-stopSignal:
//			log.Printf("Received stop signal: %v", signal)
//			run = !signal
//		default:
//		// Wait for a tick
//			<-ticker.C
//			for _, bucket := range BucketRegistry {
//				fillBucket(bucket)
//			}
//		}
//	}
//}
//
//func fillBucket(bucket *Bucket) {
//	oldTokens := bucket.tokens
//	if bucket.capacity > bucket.tokens {
//		bucket.tokens = int(math.Min(float64(bucket.capacity), float64(bucket.fillRate + bucket.tokens)))
//		log.Printf("%v received tick; tokens upped from %v to %v.", bucket.Name, oldTokens, bucket.tokens)
//	} else {
//		log.Printf("%v received tick; tokens at capacity.", bucket.Name)
//	}
//}
//
//
