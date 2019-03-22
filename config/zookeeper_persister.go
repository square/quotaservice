// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package config

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/square/quotaservice/logging"
	pb "github.com/square/quotaservice/protos/config"
)

const (
	sessionTimeout = 3 * time.Second
	createRetries  = 3
)

type eventListener func() (<-chan zk.Event, error)

type zkWatch struct {
	listener eventListener
	channel  <-chan zk.Event
	stopper  chan struct{}
}

type ZkConfigPersister struct {
	initialized bool
	// Current configuration hash
	config string

	// Historical map of configurations
	// hash -> config
	configs map[string]*pb.ServiceConfig

	// Base Zookeeper path
	path string

	watcher chan struct{}

	conn  *zk.Conn
	watch *zkWatch

	wg sync.WaitGroup
	sync.RWMutex
}

// Mirrors go-zookeeper's connOption
type connOption func(c *zk.Conn)

func NewZkConfigPersisterWithConnection(path string, conn *zk.Conn) (*ZkConfigPersister, error) {
	err := createPath(conn, path)

	if err != nil {
		conn.Close()
		return nil, err
	}

	persister := &ZkConfigPersister{
		conn:    conn,
		path:    path,
		watcher: make(chan struct{}, 1),
		configs: make(map[string]*pb.ServiceConfig)}

	watch, err := persister.createWatch(persister.currentConfigEventListener)

	if err != nil {
		conn.Close()
		return nil, err
	}

	persister.watch = watch

	return persister, nil
}

func NewZkConfigPersister(path string, servers []string, options ...connOption) (*ZkConfigPersister, error) {
	conn, _, err := zk.Connect(servers, sessionTimeout, func(c *zk.Conn) {
		c.SetLogger(logging.CurrentLogger())

		// Allows overriding logger/etc. go-zookeeper options
		for _, option := range options {
			option(c)
		}
	})

	if err != nil {
		return nil, err
	}

	return NewZkConfigPersisterWithConnection(path, conn)
}

func (z *ZkConfigPersister) createWatch(listener eventListener) (*zkWatch, error) {
	watch := &zkWatch{
		listener: listener,
		stopper:  make(chan struct{}, 1)}

	channel, err := listener()

	if err != nil {
		return nil, err
	}

	watch.channel = channel

	z.wg.Add(1)
	go z.waitForEvents(watch)

	return watch, nil
}

func (z *ZkConfigPersister) waitForEvents(watch *zkWatch) {
	defer z.wg.Done()

	for {
		select {
		case event := <-watch.channel:
			if event.Err != nil {
				logging.Print("Received error from zookeeper", event)
			} else {
				logging.Printf("Received event %+v on zookeeper watch", event)
			}
		case <-watch.stopper:
			logging.Print("Received stop signal; stopping zookeeper watcher goroutine")
			return
		}

		channel, err := watch.listener()

		if err != nil {
			logging.Printf("Received error from zookeeper executing listener: %+v", err)
			continue
		}

		watch.channel = channel
	}
}

// Tries to create the configuration path, if it doesn't exist
// It tries multiple times in case there's a race with another quotaservice node coming up
func createPath(conn *zk.Conn, path string) (err error) {
	for i := 0; i < createRetries; i++ {
		exists, _, err := conn.Exists(path)

		if exists && err == nil {
			return nil
		}

		_, err = conn.Create(path, []byte{}, 0, zk.WorldACL(zk.PermAll))

		if err == nil {
			return nil
		}

		logging.Printf("Could not create zk path, sleeping for 100ms error=%s", err.Error())
		time.Sleep(100 * time.Millisecond)
	}

	if err == nil {
		err = errors.New("could not create and get path " + path)
	}

	return err
}

// PersistAndNotify persists a configuration passed in.
func (z *ZkConfigPersister) PersistAndNotify(oldHash string, cfg *pb.ServiceConfig) error {
	// TODO(manik) Optimistic version check with oldHash

	b, e := proto.Marshal(cfg)
	if e != nil {
		return e
	}

	z.RLock()
	defer z.RUnlock()

	key := HashConfigBytes(b)
	if key == z.config {
		return nil
	}

	path := fmt.Sprintf("%s/%s", z.path, key)
	logging.Printf("Storing config version %v in path %v", cfg.Version, path)

	if err := z.archiveConfig(path, b); err != nil {
		return err
	}

	_, err := z.conn.Set(z.path, []byte(key), -1)

	// There is no notification, that happens when zookeeper alerts the watcher

	return err
}

// ReadPersistedConfig provides a config previously persisted.
func (z *ZkConfigPersister) ReadPersistedConfig() (*pb.ServiceConfig, error) {
	z.RLock()
	defer z.RUnlock()

	return CloneConfig(z.configs[z.config]), nil
}

func (z *ZkConfigPersister) currentConfigEventListener() (<-chan zk.Event, error) {
	if z.initialized {
		logging.Printf("Re-establishing zookeeper watch on %v", z.path)
	} else {
		logging.Printf("Establishing zookeeper watch on %v", z.path)
	}

	// Ignoring the response to getting the contents of the watch. We don't care which node triggered the watch, since
	// we're working out the "most recent" config below by iterating over all available configurations and sorting by
	// the configuration's Version field. All we care about here is the channel watching the ZK path, to be notified of
	// future changes.
	// Placing this before the z.conn.Children() ensures we won't miss any new versions that are created during the
	// search
	_, _, ch, err := z.conn.GetW(z.path)

	if err != nil {
		logging.Printf("Received error from zookeeper when fetching %s: %+v", z.path, err)
		return nil, err
	}

	if z.initialized {
		logging.Print("Refreshing configs from zookeeper")
	} else {
		logging.Print("Reading configs from zookeeper for the first time")
	}

	children, _, err := z.conn.Children(z.path)

	if err != nil {
		logging.Printf("Received error from zookeeper when fetching children of %s: %+v", z.path, err)
		return nil, err
	}

	configs := make(map[string]*pb.ServiceConfig)
	latestHashVersion := int32(0)
	var latestHash string

	// Iterate over all children in this path for 2 reasons: add all of them to the historical version map, and to work
	// out which is the most recent, to use as the current version.
	for _, hash := range children {
		path := fmt.Sprintf("%s/%s", z.path, hash)
		data, _, err := z.conn.Get(path)

		if err != nil {
			logging.Printf("Received error from zookeeper when fetching %s: %+v", path, err)
			return nil, err
		}

		configs[hash] = &pb.ServiceConfig{}
		proto.Unmarshal(data, configs[hash])

		// TODO(manik) replace this with a ZK node that tracks the latest version rather than deserializing each node
		// TODO(manik) we can currently have multiple hashes with the same version; this needs to be fixed at the time
		// of writing
		if configs[hash].Version >= latestHashVersion {
			latestHashVersion = configs[hash].Version
			latestHash = hash
		}
	}

	z.Lock()
	defer z.Unlock()

	z.configs = configs
	z.config = latestHash

	logging.Printf("Setting latest config hash to %v (version %v)", z.config, latestHashVersion)

	select {
	case z.watcher <- struct{}{}:
		// Notified
	default:
		// Doesn't matter; another notification is pending.
	}

	z.initialized = true

	return ch, nil
}

func (z *ZkConfigPersister) archiveConfig(path string, config []byte) error {
	_, err := z.conn.Create(path, config, 0, zk.WorldACL(zk.PermAll))
	return err
}

// ConfigChangedWatcher returns a channel that is notified whenever configuration changes are
// detected. Changes are coalesced so that a single notification may be emitted for multiple
// changes.
func (z *ZkConfigPersister) ConfigChangedWatcher() <-chan struct{} {
	return z.watcher
}

// ReadHistoricalConfigs returns an array of previously persisted configs
func (z *ZkConfigPersister) ReadHistoricalConfigs() ([]*pb.ServiceConfig, error) {
	z.RLock()
	defer z.RUnlock()

	return CloneConfigs(z.configs), nil
}

// Close makes sure all event listeners are done
// and then closes the connection
func (z *ZkConfigPersister) Close() {
	z.watch.stopper <- struct{}{}
	z.wg.Wait()

	close(z.watch.stopper)
	close(z.watcher)

	z.conn.Close()
}
