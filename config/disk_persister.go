// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

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

	d := &DiskConfigPersister{location, make(chan struct{}, 1)}

	// Notify that we're available for reading
	d.watcher <- struct{}{}

	return d, nil
}

func writeFile(path string, bytes []byte) error {
	f, e := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)

	if e != nil {
		return e
	}

	f.Write(bytes)
	return f.Close()
}

// PersistAndNotify persists a marshalled configuration passed in.
func (d *DiskConfigPersister) PersistAndNotify(marshalledConfig io.Reader) error {
	b, e := ioutil.ReadAll(marshalledConfig)

	if e != nil {
		return e
	}

	path := fmt.Sprintf("%s-%s", d.location, hashConfig(b))
	e = writeFile(path, b)

	if e != nil {
		return e
	}

	if _, e := os.Stat(d.location); e == nil {
		e = os.Remove(d.location)

		if e != nil {
			return e
		}
	}

	e = os.Symlink(path, d.location)

	if e != nil {
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
func (d *DiskConfigPersister) ReadPersistedConfig() (io.Reader, error) {
	return os.Open(d.location)
}

// ReadHistoricalConfigs returns an array of previously persisted configs
func (d *DiskConfigPersister) ReadHistoricalConfigs() ([]io.Reader, error) {
	files, err := filepath.Glob(fmt.Sprintf("%s-*", d.location))

	if err != nil {
		return nil, err
	}

	configs := make([]io.Reader, len(files))

	for i, file := range files {
		reader, e := os.Open(file)

		if e != nil {
			return nil, e
		}

		configs[i] = reader
	}

	return configs, nil
}

// ConfigChangedWatcher returns a channel that is notified whenever configuration changes are
// detected. Changes are coalesced so that a single notification may be emitted for multiple
// changes.
func (d *DiskConfigPersister) ConfigChangedWatcher() <-chan struct{} {
	return d.watcher
}
