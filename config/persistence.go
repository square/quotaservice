// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package config

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
)

// ConfigPersister is an interface that persists configs and notifies a channel of changes.
type ConfigPersister interface {
	// PersistAndNotify persists a marshalled configuration passed in.
	PersistAndNotify(marshalledConfig io.Reader) error
	// ConfigChangedWatcher returns a channel that is notified whenever configuration changes are
	// detected. Changes are coalesced so that a single notification may be emitted for multiple
	// changes.
	ConfigChangedWatcher() chan struct{}
	// ReadPersistedConfig provides a reader to a marshalled config previously persisted.
	ReadPersistedConfig() (marshalledConfig io.Reader, err error)
}

// DiskConfigPersister is a ConfigPersister that saves configs to the local filesystem.
type DiskConfigPersister struct {
	location string
	watcher  chan struct{}
}

// NewDiskConfigPersister creates a new DiskConfigPersister
func NewDiskConfigPersister(location string) (ConfigPersister, error) {
	_, e := os.Stat(location)
	// This will catch nonexistent paths, as well as passing in a directory instead of a file.
	// Nonexistent files in an existing path, however, is allowed.
	if e != nil && !os.IsNotExist(e) {
		return nil, e
	}

	return &DiskConfigPersister{location, make(chan struct{}, 1)}, nil
}

// PersistAndNotify persists a marshalled configuration passed in.
func (d *DiskConfigPersister) PersistAndNotify(marshalledConfig io.Reader) error {
	f, e := os.OpenFile(d.location, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if e != nil {
		return e
	}

	b, e := ioutil.ReadAll(marshalledConfig)
	if e != nil {
		return e
	}

	f.Write(b)
	if e = f.Close(); e != nil {
		return e
	}

	// ... and notify
	select {
	case d.watcher <- struct{}{}:
		// Notified
	default:
		// Doesn't matter; another notification is pending.
	}
	return nil
}

// ReadPersistedConfig provides a reader to a marshalled config previously persisted.
func (d *DiskConfigPersister) ReadPersistedConfig() (marshalledConfig io.Reader, err error) {
	b, e := ioutil.ReadFile(d.location)
	if e != nil {
		return nil, e
	}

	return bytes.NewReader(b), nil
}

// ConfigChangedWatcher returns a channel that is notified whenever configuration changes are
// detected. Changes are coalesced so that a single notification may be emitted for multiple
// changes.
func (d *DiskConfigPersister) ConfigChangedWatcher() chan struct{} {
	return d.watcher
}
