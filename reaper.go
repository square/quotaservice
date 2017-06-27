// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package quotaservice

import (
	"time"

	"github.com/square/quotaservice/config"
	"github.com/square/quotaservice/logging"
	pbconfig "github.com/square/quotaservice/protos/config"
)

// watcher watches reapableBuckets for activity.
type watcher struct {
	ns           string
	bucketName   string
	identifier   string
	maxIdle      time.Duration
	lastActivity time.Time
	activities   <-chan struct{}
}

// activityDetected tells you if activity has been detected since the last time this method was
// called.
func (w *watcher) activityDetected() bool {
	select {
	case <-w.activities:
		return true
	default:
		return false
	}
}

// tooIdle returns true if a bucket has been idle for longer than its maxIdle
func (w *watcher) tooIdle(now time.Time) bool {
	if w.activityDetected() {
		w.lastActivity = now
		return false
	}
	return now.Sub(w.lastActivity) > w.maxIdle
}

// reapableBucket is a wrapper around a bucket that overrides ReportActivity(), and reports any
// activity on a channel that is monitored by the reaper.
type reapableBucket struct {
	Bucket
	activities chan<- struct{}
}

// ReportActivity is overridden to report activity to the reapableBucket's activity channel.
func (r *reapableBucket) ReportActivity() {
	select {
	case r.activities <- struct{}{}:
	// reported activity
	default:
		// Already reported
	}

	// Now call ReportActivity() on the delegate bucket.
	r.Bucket.ReportActivity()
}

// Destroy closes the reapableBucket's activity channel.
func (r *reapableBucket) Destroy() {
	close(r.activities)

	// Now call Destroy() on the delegate bucket.
	r.Bucket.Destroy()
}

func createWatcher(ns, bucketName string, maxIdle time.Duration, activityChannel <-chan struct{}) *watcher {
	return &watcher{
		ns:         ns,
		bucketName: bucketName,
		identifier: config.FullyQualifiedName(ns, bucketName),
		maxIdle:    maxIdle,
		activities: activityChannel}
}

type reaper struct {
	cfg         config.ReaperConfig
	newWatchers chan<- *watcher
	watchers    map[string]*watcher
}

func newReaper(bc *bucketContainer, r config.ReaperConfig) *reaper {
	watcherChannel := make(chan *watcher, r.BucketWatcherBuffer)
	reaper := &reaper{
		cfg:         r,
		watchers:    make(map[string]*watcher),
		newWatchers: watcherChannel}

	go reaper.reapIdleBuckets(bc, watcherChannel)

	return reaper
}

func (r *reaper) stop() {
	// This will trigger the goroutine waiting on watchers to exit.
	close(r.newWatchers)
}

func (r *reaper) addNewWatcher(w *watcher) {
	r.watchers[w.identifier] = w
	w.lastActivity = time.Now()
}

// checkExpirations checks all watches registered with the reaper, and destroys idle buckets, updating the reaper
// accordingly. Returns the duration after which it should run again.
func (r *reaper) checkExpirations(bc *bucketContainer) time.Duration {
	now := time.Now()
	newSleep := r.cfg.MinFrequency
	var reaped uint64
	for id, w := range r.watchers {
		if w.tooIdle(now) {
			// Reap bucket
			reaped++
			if bc.removeBucket(w.ns, w.bucketName) {
				delete(r.watchers, id)
			}
		} else if w.maxIdle < newSleep {
			// Check if we're sleeping the right amount.
			newSleep = w.maxIdle
		}
	}
	logging.Printf("Reaped %d buckets due to inactivity", reaped)
	return newSleep
}

// watch watches all buckets for activity, deleting the bucket if no activity has been detected
// after a given duration.
func (r *reaper) reapIdleBuckets(bc *bucketContainer, newWatchers <-chan *watcher) {
	sleep := r.cfg.InitSleep
	logging.Printf("reapIdleBuckets started. Initial sleep %v", sleep)
	ticker := time.NewTicker(sleep)

	// Watch on a ticker, or a new watch being created.
	for {
		select {
		case w, ok := <-newWatchers:
			if ok {
				r.addNewWatcher(w)
			} else {
				// newWatchers closed; stop the reaper.
				r.newWatchers = nil
				r.watchers = nil
				return
			}

		case <-ticker.C:
			newSleep := r.checkExpirations(bc)

			if newSleep != sleep {
				logging.Printf("Adjusting ticker to run with duration %v", newSleep)
				// We need a new ticker.
				ticker.Stop()
				ticker = time.NewTicker(newSleep)
				sleep = newSleep
			}
		}
	}
}

// applyWatch decorates a bucket to make it "watchable", if it has a maxIdle and requires garbage
// collection. Callers should ensure they point to the return value of this method when
// referencing their bucket.
func (r *reaper) applyWatch(delegate Bucket, namespace, bucketName string, cfg *pbconfig.BucketConfig) (Bucket, *watcher) {
	if cfg.MaxIdleMillis > 0 {
		activityChannel := make(chan struct{}, 1)
		rb := &reapableBucket{Bucket: delegate, activities: activityChannel}
		w := createWatcher(namespace, bucketName, time.Duration(cfg.MaxIdleMillis)*time.Millisecond, activityChannel)
		r.newWatchers <- w
		return rb, w
	}

	return delegate, nil
}
