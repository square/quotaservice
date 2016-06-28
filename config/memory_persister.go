// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package config

import (
	"errors"
	"io"
)

type MemoryConfigPersister struct {
	config  io.Reader
	watcher chan struct{}
}

func NewMemoryConfigPersister() (ConfigPersister, error) {
	return &MemoryConfigPersister{nil, make(chan struct{}, 1)}, nil
}

// PersistAndNotify persists a marshalled configuration passed in.
func (m *MemoryConfigPersister) PersistAndNotify(marshalledConfig io.Reader) error {
	m.config = marshalledConfig

	// ... and notify
	select {
	case m.watcher <- struct{}{}:
		// Notified
	default:
		// Doesn't matter; another notification is pending.
	}

	return nil
}

// ReadPersistedConfig provides a reader to a marshalled config previously persisted.
func (m *MemoryConfigPersister) ReadPersistedConfig() (marshalledConfig io.Reader, err error) {
	if m.config == nil {
		return nil, errors.New("config is empty")
	}

	return m.config, nil
}

// ConfigChangedWatcher returns a channel that is notified whenever configuration changes are
// detected. Changes are coalesced so that a single notification may be emitted for multiple
// changes.
func (m *MemoryConfigPersister) ConfigChangedWatcher() chan struct{} {
	return m.watcher
}
