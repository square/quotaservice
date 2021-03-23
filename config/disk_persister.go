// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang/protobuf/proto"
	"github.com/square/quotaservice/config/internal"
	pb "github.com/square/quotaservice/protos/config"
)

// DiskConfigPersister is a ConfigPersister that saves configs to the local filesystem.
type DiskConfigPersister struct {
	location string
	*internal.Notifier
}

// NewDiskConfigPersister creates a new DiskConfigPersister
func NewDiskConfigPersister(location string) (*DiskConfigPersister, error) {
	_, e := os.Stat(location)
	// This will catch nonexistent paths, as well as passing in a directory instead of a file.
	// Nonexistent files in an existing path, however, is allowed.
	if e != nil && !os.IsNotExist(e) {
		return nil, e
	}

	d := &DiskConfigPersister{location, internal.NewNotifier()}

	// Notify that we're available for reading
	d.Notify()

	return d, nil
}

func writeFile(path string, bytes []byte) error {
	f, e := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if e != nil {
		return e
	}

	defer func() { _ = f.Close() }()

	_, e = f.Write(bytes)
	return e
}

// PersistAndNotify persists a configuration passed in.
func (d *DiskConfigPersister) PersistAndNotify(oldHash string, cfg *pb.ServiceConfig) error {
	// TODO(manik) Optimistic version check with oldHash

	b, e := proto.Marshal(cfg)
	if e != nil {
		return e
	}

	path := fmt.Sprintf("%s-%s", d.location, HashConfigBytes(b))
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
	d.Notify()

	return nil
}

// ReadPersistedConfig provides a config previously persisted.
func (d *DiskConfigPersister) ReadPersistedConfig() (*pb.ServiceConfig, error) {
	r, e := os.Open(d.location)
	if e != nil {
		return nil, e
	}

	return Unmarshal(r)
}

// ReadHistoricalConfigs returns an array of previously persisted configs
func (d *DiskConfigPersister) ReadHistoricalConfigs() ([]*pb.ServiceConfig, error) {
	files, err := filepath.Glob(fmt.Sprintf("%s-*", d.location))

	if err != nil {
		return nil, err
	}

	configs := make([]*pb.ServiceConfig, len(files))

	for i, file := range files {
		reader, e := os.Open(file)
		if e != nil {
			return nil, e
		}

		configs[i], e = Unmarshal(reader)
		if e != nil {
			return nil, e
		}
	}

	return configs, nil
}

// ConfigChangedWatcher returns a channel that is notified whenever configuration changes are
// detected. Changes are coalesced so that a single notification may be emitted for multiple
// changes.
func (d *DiskConfigPersister) ConfigChangedWatcher() <-chan struct{} {
	return d.Notifier.Watcher
}

// Close closes the notification channel.
func (d *DiskConfigPersister) Close() {
	close(d.Notifier.Watcher)
}
