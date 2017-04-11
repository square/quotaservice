// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package config

import (
	"bytes"
	"io"
	"io/ioutil"
)

type MemoryConfigPersister struct {
	config  string
	configs map[string][]byte
	*Notifier
}

func NewMemoryConfigPersister() ConfigPersister {
	p := &MemoryConfigPersister{
		configs:  make(map[string][]byte),
		Notifier: NewNotifier()}

	p.Notify()
	return p
}

// PersistAndNotify persists a marshalled configuration passed in.
func (m *MemoryConfigPersister) PersistAndNotify(marshalledConfig io.Reader) error {
	b, err := ioutil.ReadAll(marshalledConfig)
	if err != nil {
		return err
	}

	m.config = HashConfig(b)
	m.configs[m.config] = b

	// ... and notify
	m.Notify()

	return nil
}

// ReadPersistedConfig provides a reader to a marshalled config previously persisted.
func (m *MemoryConfigPersister) ReadPersistedConfig() (io.Reader, error) {
	return bytes.NewReader(m.configs[m.config]), nil
}

// ReadHistoricalConfigs returns an array of previously persisted configs
func (m *MemoryConfigPersister) ReadHistoricalConfigs() ([]io.Reader, error) {
	var readers []io.Reader

	for _, v := range m.configs {
		readers = append(readers, bytes.NewReader(v))
	}

	return readers, nil
}

// ConfigChangedWatcher returns a channel that is notified whenever configuration changes are
// detected. Changes are coalesced so that a single notification may be emitted for multiple
// changes.
func (m *MemoryConfigPersister) ConfigChangedWatcher() <-chan struct{} {
	return m.Notifier.Watcher
}
