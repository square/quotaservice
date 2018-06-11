// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/square/quotaservice/config"
	"github.com/square/quotaservice/config/cloudpersisters/google"
	pb "github.com/square/quotaservice/protos/config"
	"github.com/square/quotaservice/test/helpers"
)

// Change these constants to match your settings on Google Cloud. Visit https://console.cloud.google.com/datastore
// for more details.
const (
	projectId       = "cloud-hackweek-20180611"
	credentialsFile = "/Users/manik/cloud-hackweek-20180611.json"

	// WARNING: DO NOT point to a production namespace, since this will write test data.
	dsNamespace = "QuotaServiceConfigs"
	initVersion = 0
)

var counter = 0

// main runs a "manual" test, which involves setting up your credentials and Google Cloud account details in the
// constants above. As such, this isn't designed to be run in CI.
//
// WARNING: DO NOT point to a production namespace, since this will write test data.
func main() {
	dp := createGoogleDatastorePersister()
	configChangedWatcher := dp.ConfigChangedWatcher()
	consumeAll(configChangedWatcher)

	// Create an initial configuration to persist.
	cfg := config.NewDefaultServiceConfig()
	cfg.Version = initVersion
	updateConfig(cfg)

	// Persist the configuration.
	e := dp.PersistAndNotify("", cfg) // TODO use a real oldHash to test optimistic version check
	helpers.PanicError(e)
	consume(configChangedWatcher, 10*time.Second)

	// Modify configuration.
	updateConfig(cfg)

	// Persist again.
	e = dp.PersistAndNotify("", cfg)
	helpers.PanicError(e)
	consume(configChangedWatcher, 10*time.Second)

	time.Sleep(time.Second * 10)

	// Modify configuration.
	updateConfig(cfg)

	// Persist again.
	e = dp.PersistAndNotify("", cfg)
	helpers.PanicError(e)
	consume(configChangedWatcher, 10*time.Second)

	time.Sleep(time.Second * 10)

	cfgs, e := dp.ReadHistoricalConfigs()
	helpers.PanicError(e)

	for i, c := range cfgs {
		fmt.Printf("Cfg number %v:\n%+v\n\n", i, c)
	}

	cfg, e = dp.ReadPersistedConfig()
	helpers.PanicError(e)

	helpers.PanicError(e)
	fmt.Printf("Latest cfg:\n%+v\n\n", cfg)
}

func createGoogleDatastorePersister() config.ConfigPersister {
	dp, e := google.New(projectId, credentialsFile, dsNamespace, time.Second)
	helpers.PanicError(e)
	return dp
}

// consumeAll consumes all messages on a channel until there is nothing left.
func consumeAll(c <-chan struct{}) {
	for {
		select {
		case <-c:
			// keep looping
		default:
			return
		}
	}
}

// consume consumes a single element from a channel, blocking until it does so.
func consume(c <-chan struct{}, maxWait time.Duration) {
	t := time.NewTimer(maxWait)
	select {
	case <-c:
		return
	case <-t.C:
		panic(fmt.Sprintf("Timed out waiting for an event for %v", maxWait))
	}
}

func updateConfig(cfg *pb.ServiceConfig) {
	cfg.GlobalDefaultBucket = config.NewDefaultBucketConfig(config.DefaultBucketName)
	cfg.GlobalDefaultBucket.FillRate = rand.Int63()
	counter++
	cfg.User = fmt.Sprintf("user-%v", counter)
	cfg.Date = time.Now().Unix()
	cfg.Version += 1
}
