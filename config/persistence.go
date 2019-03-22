// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package config

import (
	"crypto/md5"
	"fmt"
	"github.com/square/quotaservice/logging"
	pb "github.com/square/quotaservice/protos/config"
	"io/ioutil"
)

// ConfigPersister is an interface that persists configs and notifies a channel of changes.
type ConfigPersister interface {
	// PersistAndNotify persists a configuration passed in.
	PersistAndNotify(oldHash string, newConfig *pb.ServiceConfig) error
	// ConfigChangedWatcher returns a channel that is notified whenever configuration changes are
	// detected. Changes are coalesced so that a single notification may be emitted for multiple
	// changes.
	ConfigChangedWatcher() <-chan struct{}
	// ReadPersistedConfig provides a config previously persisted.
	ReadPersistedConfig() (*pb.ServiceConfig, error)
	// Returns an array of historical configurations, used to display a history for admin consoles.
	ReadHistoricalConfigs() ([]*pb.ServiceConfig, error)
}

// HashConfigBytes returns the MD5 of a config byte array.
func HashConfigBytes(cfgBytes []byte) string {
	return fmt.Sprintf("%x", md5.Sum(cfgBytes))
}

// HashConfig returns the MD5 of a service config
func HashConfig(config *pb.ServiceConfig) string {
	r, err := Marshal(config)

	if err != nil {
		logging.Printf("Unable to marshal config %+v: %v", config, err)
		return ""
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		logging.Print("Unable to read bytes", err)
		return ""
	}

	return HashConfigBytes(b)
}
