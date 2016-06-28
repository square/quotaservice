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
	SessionTimeout = 3 * time.Second
)

type ZkConfigPersister struct {
	conn    *zk.Conn
	path    string
	config  io.Reader
	watcher chan struct{}
	wg      sync.WaitGroup
}

func NewZkConfigPersister(path string, servers []string) (ConfigPersister, error) {
	conn, _, err := zk.Connect(servers, SessionTimeout)

	if err != nil {
		return nil, err
	}

	conf, ch, err := createAndSetWatch(conn, path)

	if err != nil {
		conn.Close()
		return nil, err
	}

	persister := &ZkConfigPersister{
		conn:    conn,
		path:    path,
		watcher: make(chan struct{}, 1)}

	persister.setAndNotify(conf)

	persister.wg.Add(1)
	go persister.zkEventListener(ch)

	return persister, nil
}

// Sets a watch on the given path. If the path does not exist, it tries to create it.
// However, it tries multiple times in case there's a race with another quotaservice node coming up
func createAndSetWatch(conn *zk.Conn, path string) ([]byte, <-chan zk.Event, error) {
	var err error

	for i := 0; i < 3; i++ {
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

		conf, _, ch, err := conn.GetW(path)

		if err == nil {
			return conf, ch, nil
		}

		logging.Printf("Could not set zk watch, sleeping for 100ms")
		time.Sleep(100 * time.Millisecond)
	}

	if err == nil {
		err = errors.New("could not create and get path " + path)
	}

	return nil, nil, err
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
func (z *ZkConfigPersister) ReadPersistedConfig() (marshalledConfig io.Reader, err error) {
	return z.config, nil
}

func (z *ZkConfigPersister) zkEventListener(ch <-chan zk.Event) {
	for event := range ch {
		if event.Err != nil {
			logging.Print("Received error from zookeeper", event)
			continue
		}

		if event.Type != zk.EventNodeDataChanged {
			continue
		}

		config, _, err := z.conn.Get(z.path)

		if err != nil {
			logging.Printf("Received error from zookeeper when fetching %s: %+v", event, z.path)
			continue
		}

		z.setAndNotify(config)
	}

	z.wg.Done()
}

func (z *ZkConfigPersister) setAndNotify(config []byte) {
	readerConfig := bytes.NewReader(config)
	z.config = readerConfig

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
	close(z.watcher)
	z.conn.Close()
	z.wg.Wait()
}
