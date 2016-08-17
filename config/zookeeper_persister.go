// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package config

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/maniksurtani/quotaservice/logging"
	"github.com/samuel/go-zookeeper/zk"
)

const (
	sessionTimeout = 3 * time.Second
	watchRetries   = 3
)

type ZkConfigPersister struct {
	conn    *zk.Conn
	path    string
	config  []byte
	watcher chan struct{}
	stopper chan struct{}
	wg      sync.WaitGroup
}

func NewZkConfigPersister(path string, servers []string) (ConfigPersister, error) {
	conn, _, err := zk.Connect(servers, sessionTimeout)

	if err != nil {
		return nil, err
	}

	conf, err := createAndGetConfig(conn, path)

	if err != nil {
		conn.Close()
		return nil, err
	}

	persister := &ZkConfigPersister{
		conn:    conn,
		path:    path,
		watcher: make(chan struct{}, 1),
		stopper: make(chan struct{}, 1)}

	persister.setAndNotify(conf)

	persister.wg.Add(1)
	go persister.zkEventListener()

	return persister, nil
}

// If the path does not exist, it tries to create it.
// However, it tries multiple times in case there's a race with another quotaservice node coming up
func createAndGetConfig(conn *zk.Conn, path string) ([]byte, error) {
	var err error

	for i := 0; i < watchRetries; i++ {
		exists, _, err := conn.Exists(path)

		if err != nil {
			continue
		}

		if !exists {
			_, err = conn.Create(path, []byte{}, 0, zk.WorldACL(zk.PermAll))

			if err != nil {
				continue
			}
		}

		conf, _, err := conn.Get(path)

		if err == nil {
			return conf, nil
		}

		logging.Printf("Could not get zk config, sleeping for 100ms")
		time.Sleep(100 * time.Millisecond)
	}

	if err == nil {
		err = errors.New("could not create and get path " + path)
	}

	return nil, err
}

// PersistAndNotify persists a marshalled configuration passed in.
func (z *ZkConfigPersister) PersistAndNotify(marshalledConfig io.Reader) error {
	b, e := ioutil.ReadAll(marshalledConfig)
	if e != nil {
		return e
	}

	_, err := z.conn.Set(z.path, b, -1)

	// There is no notification, that happens when zookeeper alerts the watcher

	return err
}

// ReadPersistedConfig provides a reader to a marshalled config previously persisted.
func (z *ZkConfigPersister) ReadPersistedConfig() (io.Reader, error) {
	return bytes.NewReader(z.config), nil
}

func (z *ZkConfigPersister) zkEventListener() {
	for {
		select {
		case <-z.stopper:
			z.wg.Done()
			return
		default:
		}

		config, _, ch, err := z.conn.GetW(z.path)

		if err != nil {
			logging.Printf("Received error from zookeeper when fetching %s: %+v", z.path, err)
			// TODO(@steved) backoff?
			continue
		}

		z.setAndNotify(config)

		event := <-ch

		if event.Err != nil {
			logging.Printf("Received error from zookeeper: %+v", event)
		}
	}
}

func (z *ZkConfigPersister) setAndNotify(config []byte) {
	z.config = config

	// ... and notify
	select {
	case z.watcher <- struct{}{}:
		// Notified
	default:
		// Doesn't matter; another notification is pending.
	}
}

// ConfigChangedWatcher returns a channel that is notified whenever configuration changes are
// detected. Changes are coalesced so that a single notification may be emitted for multiple
// changes.
func (z *ZkConfigPersister) ConfigChangedWatcher() chan struct{} {
	return z.watcher
}

func (z *ZkConfigPersister) Close() {
	z.stopper <- struct{}{}
	z.conn.Close()
	z.wg.Wait()
	close(z.watcher)
	close(z.stopper)
}
