// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package main

import (
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/square/quotaservice/config"
	"github.com/square/quotaservice/config/cloudpersisters/google"
	"github.com/square/quotaservice/protos/config"
)

// Change these constants to match your settings on Google Cloud. Visit https://console.cloud.google.com/datastore
// for more details.
const (
	projectId       = "my-project-id-on-google-cloud"
	credentialsFile = "/path/to/my/credentials/file.json"
	namespace       = "MyGoogleDatastoreNamespace"
	entity          = "MyGoogleDatastoreEntityKind"
	initVersion     = 0
)

var counter = 0

// main runs a "manual" test, which involves setting up your credentials and Google Cloud account details in the
// constants above. As such, this isn't designed to be run in CI.
func main() {
	dp := createGoogleDatastorePersister()
	configChangedWatcher := dp.ConfigChangedWatcher()
	consumeAll(configChangedWatcher)

	// Create an initial configuration to persist.
	cfg := config.NewDefaultServiceConfig()
	cfg.Version = initVersion
	r := updateConfig(cfg)

	// Persist the configuration.
	e := dp.PersistAndNotify(r)
	checkNoErrors(e)
	consume(configChangedWatcher, 10*time.Second)

	// Modify configuration.
	r = updateConfig(cfg)

	// Persist again.
	e = dp.PersistAndNotify(r)
	checkNoErrors(e)
	consume(configChangedWatcher, 10*time.Second)

	time.Sleep(time.Second * 10)

	// Modify configuration.
	r = updateConfig(cfg)

	// Persist again.
	e = dp.PersistAndNotify(r)
	checkNoErrors(e)
	consume(configChangedWatcher, 10*time.Second)

	time.Sleep(time.Second * 10)

	cfgs, e := dp.ReadHistoricalConfigs()
	checkNoErrors(e)

	for i, c := range cfgs {
		cfg, e := config.Unmarshal(c)
		checkNoErrors(e)
		fmt.Printf("Cfg number %v:\n%+v\n\n", i, cfg)
	}

	c, e := dp.ReadPersistedConfig()
	checkNoErrors(e)

	cfg, e = config.Unmarshal(c)
	checkNoErrors(e)
	fmt.Printf("Latest cfg:\n%+v\n\n", cfg)
}

func createGoogleDatastorePersister() config.ConfigPersister {
	dp, e := google.New(projectId, credentialsFile, namespace, entity, time.Second)
	checkNoErrors(e)
	return dp
}

func checkNoErrors(e error) {
	if e != nil {
		panic(e)
	}
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

func updateConfig(cfg *quotaservice_configs.ServiceConfig) io.Reader {
	cfg.GlobalDefaultBucket = config.NewDefaultBucketConfig(config.DefaultBucketName)
	cfg.GlobalDefaultBucket.FillRate = rand.Int63()
	counter++
	cfg.User = fmt.Sprintf("user-%v", counter)
	cfg.Date = time.Now().Unix()
	cfg.Version += 1

	r, e := config.Marshal(cfg)
	checkNoErrors(e)
	return r
}
