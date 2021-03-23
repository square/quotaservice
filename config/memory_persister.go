// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package config

import (
	"sync"

	"github.com/square/quotaservice/config/internal"
	pb "github.com/square/quotaservice/protos/config"
)

type MemoryConfigPersister struct {
	config  string
	configs map[string]*pb.ServiceConfig
	*internal.Notifier
	*sync.RWMutex
}

func NewMemoryConfigPersister() *MemoryConfigPersister {
	p := &MemoryConfigPersister{
		configs:  make(map[string]*pb.ServiceConfig),
		Notifier: internal.NewNotifier(),
		RWMutex:  &sync.RWMutex{}}

	p.Notify()
	return p
}

// PersistAndNotify persists a configuration passed in.
func (m *MemoryConfigPersister) PersistAndNotify(oldHash string, cfg *pb.ServiceConfig) error {
	// TODO(manik) Optimistic version check with oldHash

	m.Lock()
	defer m.Unlock()

	m.config = HashConfig(cfg)
	m.configs[m.config] = CloneConfig(cfg)

	// ... and notify
	m.Notify()

	return nil
}

// ReadPersistedConfig provides a config previously persisted.
func (m *MemoryConfigPersister) ReadPersistedConfig() (*pb.ServiceConfig, error) {
	m.RLock()
	defer m.RUnlock()

	return CloneConfig(m.configs[m.config]), nil
}

// ReadHistoricalConfigs returns an array of previously persisted configs
func (m *MemoryConfigPersister) ReadHistoricalConfigs() ([]*pb.ServiceConfig, error) {
	m.RLock()
	defer m.RUnlock()

	return CloneConfigs(m.configs), nil
}

// ConfigChangedWatcher returns a channel that is notified whenever configuration changes are
// detected. Changes are coalesced so that a single notification may be emitted for multiple
// changes.
func (m *MemoryConfigPersister) ConfigChangedWatcher() <-chan struct{} {
	return m.Notifier.Watcher
}

// Close closes the notification channel.
func (m *MemoryConfigPersister) Close() {
	close(m.Notifier.Watcher)
}
