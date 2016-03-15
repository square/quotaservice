// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package admin

import (
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	pb "github.com/maniksurtani/quotaservice/protos/config"
)

// ConfigPersister blah blah
type ConfigPersister interface {
	PersistAndNotify(c *pb.ServiceConfig) error
	ConfigChangedWatcher() chan struct{}
	ReadPersistedConfig() (*pb.ServiceConfig, error)
}

type DiskConfigPersister struct {
	location string
	watcher  chan struct{}
}

func NewDiskConfigPersister(location string) (ConfigPersister, error) {
	_, e := os.Stat(location)
	if e != nil && !os.IsNotExist(e) {
		// TODO(manik) is this good enough to catch directories that don't exist?
		return nil, e
	}

	// TODO(manik) test that the location is writeable.

	return &DiskConfigPersister{location, make(chan struct{}, 1)}, nil
}

func (d *DiskConfigPersister) PersistAndNotify(c *pb.ServiceConfig) error {
	// Write to disk
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

func (d *DiskConfigPersister) ConfigChangedWatcher() chan struct{} {
	return d.watcher
}
