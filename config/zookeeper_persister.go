// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/maniksurtani/quotaservice/logging"
	"github.com/samuel/go-zookeeper/zk"
)

const (
	sessionTimeout = 3 * time.Second
	createRetries  = 3
)

type eventListener func() (<-chan zk.Event, error)

type ZkWatch struct {
	listener eventListener
	channel  <-chan zk.Event
	stopper  chan struct{}
}

type ZkConfigPersister struct {
	// Current configuration hash
	config string

	// Historical map of configurations
	// hash -> marshalled config
	configs map[string][]byte

	// Base Zookeeper path
	path string

	watcher chan struct{}

	conn  *zk.Conn
	watch *ZkWatch

	wg sync.WaitGroup
	sync.RWMutex
}

// Mirrors go-zookeeper's connOption
type connOption func(c *zk.Conn)

func NewZkConfigPersister(path string, servers []string, options ...connOption) (ConfigPersister, error) {
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

	err = createPath(conn, path)

	if err != nil {
		conn.Close()
		return nil, err
	}

	persister := &ZkConfigPersister{
		conn:    conn,
		path:    path,
		watcher: make(chan struct{}, 1),
		configs: make(map[string][]byte)}

	watch, err := persister.createWatch(persister.currentConfigEventListener)

	if err != nil {
		conn.Close()
		return nil, err
	}

	persister.watch = watch

	return persister, nil
}

func (z *ZkConfigPersister) createWatch(listener eventListener) (*ZkWatch, error) {
	watch := &ZkWatch{
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

func (z *ZkConfigPersister) waitForEvents(watch *ZkWatch) {
	defer z.wg.Done()

	for {
		select {
		case event := <-watch.channel:
			if event.Err != nil {
				logging.Print("Received error from zookeeper", event)
			}
		case <-watch.stopper:
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

		logging.Printf("Could not create zk path, sleeping for 100ms")
		time.Sleep(100 * time.Millisecond)
	}

	if err == nil {
		err = errors.New("could not create and get path " + path)
	}

	return err
}

// PersistAndNotify persists a marshalled configuration passed in.
func (z *ZkConfigPersister) PersistAndNotify(marshalledConfig io.Reader) error {
	b, e := ioutil.ReadAll(marshalledConfig)

	if e != nil {
		return e
	}

	z.RLock()
	defer z.RUnlock()

	key := HashConfig(b)
	if key == z.config {
		return nil
	}

	if err := z.archiveConfig(key, b); err != nil {
		return err
	}

	_, err := z.conn.Set(z.path, []byte(key), -1)

	// There is no notification, that happens when zookeeper alerts the watcher

	return err
}

// ReadPersistedConfig provides a reader to a marshalled config previously persisted.
func (z *ZkConfigPersister) ReadPersistedConfig() (io.Reader, error) {
	z.RLock()
	defer z.RUnlock()

	return bytes.NewReader(z.configs[z.config]), nil
}

func (z *ZkConfigPersister) currentConfigEventListener() (<-chan zk.Event, error) {
	children, _, err := z.conn.Children(z.path)

	if err != nil {
		logging.Printf("Received error from zookeeper when fetching children of %s: %+v", z.path, err)
		return nil, err
	}

	configs := make(map[string][]byte)

	for _, child := range children {
		path := fmt.Sprintf("%s/%s", z.path, child)
		data, _, err := z.conn.Get(path)

		if err != nil {
			logging.Printf("Received error from zookeeper when fetching %s: %+v", path, err)
			return nil, err
		} else {
			configs[child] = data
		}
	}

	config, _, ch, err := z.conn.GetW(z.path)

	if err != nil {
		logging.Printf("Received error from zookeeper when fetching %s: %+v", z.path, err)
		return nil, err
	}

	z.Lock()
	defer z.Unlock()

	z.configs = configs
	z.config = string(config)

	select {
	case z.watcher <- struct{}{}:
		// Notified
	default:
		// Doesn't matter; another notification is pending.
	}

	return ch, nil
}

func (z *ZkConfigPersister) archiveConfig(key string, config []byte) error {
	path := fmt.Sprintf("%s/%s", z.path, key)
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
func (z *ZkConfigPersister) ReadHistoricalConfigs() ([]io.Reader, error) {
	z.RLock()
	defer z.RUnlock()

	readers := make([]io.Reader, 0)

	for _, config := range z.configs {
		readers = append(readers, bytes.NewReader(config))
	}

	return readers, nil
}

// Closes makes sure all event listeners are done
// and then closes the connection
func (z *ZkConfigPersister) Close() {
	z.watch.stopper <- struct{}{}
	z.wg.Wait()

	close(z.watch.stopper)
	close(z.watcher)

	z.conn.Close()
}
