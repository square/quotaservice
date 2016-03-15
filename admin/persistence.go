// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package admin

import (
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	pb "github.com/maniksurtani/quotaservice/protos/config"
)

// ConfigPersister is an interface that persists configs and notifies a channel of changes.
type ConfigPersister interface {
	// PersistAndNotify persists the configuration passed in, returning any errors encountered.
	PersistAndNotify(c *pb.ServiceConfig) error
	// ConfigChangedWatcher returns a channel that is notified whenever configuration changes are
	// detected. Changes are coalesced so that a single notification may be emitted for multiple
	// changes.
	ConfigChangedWatcher() chan struct{}
	// ReadPersistedConfig reads configuration previously persisted, returning the configuration 
    // read and any errors encountered.
	ReadPersistedConfig() (*pb.ServiceConfig, error)
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

// PersistAndNotify persists the configuration passed in, returning any errors encountered.
func (d *DiskConfigPersister) PersistAndNotify(c *pb.ServiceConfig) error {
	f, e := os.OpenFile(d.location, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if e != nil {
		return e
	}

	var bytes []byte
	bytes, e = proto.Marshal(c)

	if e != nil {
		return e
	}

	f.Write(bytes)
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

// ReadPersistedConfig reads configuration previously persisted, returning the configuration read and 
// any errors encountered.
func (d *DiskConfigPersister) ReadPersistedConfig() (*pb.ServiceConfig, error) {
	bytes, e := ioutil.ReadFile(d.location)
	if e != nil {
		return nil, e
	}

	c := &pb.ServiceConfig{}
	e = proto.Unmarshal(bytes, c)
	if e != nil {
		return nil, e
	}

	return c, nil
}

// ConfigChangedWatcher returns a channel that is notified whenever configuration changes are
// detected. Changes are coalesced so that a single notification may be emitted for multiple
// changes.
func (d *DiskConfigPersister) ConfigChangedWatcher() chan struct{} {
	return d.watcher
}
