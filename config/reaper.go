// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package config

import "time"

// ReaperConfig represents the configuration settings for the bucket reaper.
type ReaperConfig struct {
	BucketWatcherBuffer int
	InitSleep           time.Duration
	MinFrequency        time.Duration
}

// NewReaperConfig returns a new ReaperConfig with defaults.
func NewReaperConfig() ReaperConfig {
	return ReaperConfig{
		BucketWatcherBuffer: 10000,
		InitSleep:           10 * time.Second,
		MinFrequency:        10 * time.Minute}
}
