// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package config

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
)

type MemoryConfigPersister struct {
	config  []byte
	configs map[string][]byte
	watcher chan struct{}
}

func NewMemoryConfigPersister() ConfigPersister {
	return &MemoryConfigPersister{nil, make(map[string][]byte), make(chan struct{}, 1)}
}

// PersistAndNotify persists a marshalled configuration passed in.
func (m *MemoryConfigPersister) PersistAndNotify(marshalledConfig io.Reader) error {
	bytes, err := ioutil.ReadAll(marshalledConfig)

	if err != nil {
		return err
	}

	if m.config != nil {
		m.configs[hashConfig(m.config)] = m.config
	}

	m.config = bytes

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
func (m *MemoryConfigPersister) ReadPersistedConfig() (io.Reader, error) {
	if m.config == nil {
		return nil, errors.New("config is empty")
	}

	return bytes.NewReader(m.config), nil
}

// ReadHistoricalConfigs returns an array of previously persisted configs
func (m *MemoryConfigPersister) ReadHistoricalConfigs() ([]io.Reader, error) {
	readers := make([]io.Reader, 0)

	for _, v := range m.configs {
		readers = append(readers, bytes.NewReader(v))
	}

	return readers, nil
}

// ConfigChangedWatcher returns a channel that is notified whenever configuration changes are
// detected. Changes are coalesced so that a single notification may be emitted for multiple
// changes.
func (m *MemoryConfigPersister) ConfigChangedWatcher() chan struct{} {
	return m.watcher
}
