// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package config

import (
	"reflect"
	"testing"

	pbconfig "github.com/square/quotaservice/protos/config"
	"github.com/square/quotaservice/test/helpers"
)

func TestMemoryPersistence(t *testing.T) {
	persister := NewMemoryConfigPersister()

	select {
	case <-persister.ConfigChangedWatcher():
		// This is good.
	default:
		t.Fatal("Config channel should not be empty!")
	}

	s := &pbconfig.ServiceConfig{
		GlobalDefaultBucket: &pbconfig.BucketConfig{Size: 300, FillRate: 400, WaitTimeoutMillis: 123456},
		Namespaces:          make(map[string]*pbconfig.NamespaceConfig, 1),
		Version:             92}

	nc := &pbconfig.NamespaceConfig{
		Name:              "xyz",
		MaxDynamicBuckets: 123}

	SetDynamicBucketTemplate(nc, &pbconfig.BucketConfig{})
	helpers.CheckError(t, AddNamespace(s, nc))

	// Store s.
	r, e := Marshal(s)
	helpers.CheckError(t, e)
	e = persister.PersistAndNotify(r)
	helpers.CheckError(t, e)

	// Test notification
	select {
	case <-persister.ConfigChangedWatcher():
	// This is good.
	default:
		t.Fatal("Config channel should not be empty!")
	}

	r, e = persister.ReadPersistedConfig()
	helpers.CheckError(t, e)
	unmarshalled, e := Unmarshal(r)
	helpers.CheckError(t, e)

	if !reflect.DeepEqual(s, unmarshalled) {
		t.Fatalf("Configs should be equal! %+v != %+v", s, unmarshalled)
	}

	cfgs, e := persister.ReadHistoricalConfigs()
	helpers.CheckError(t, e)

	if len(cfgs) != 1 {
		t.Fatalf("Historical configs is not correct! %+v", cfgs)
	}

	unmarshalled, e = Unmarshal(cfgs[0])
	helpers.CheckError(t, e)

	if !reflect.DeepEqual(s, unmarshalled) {
		t.Fatalf("Configs should be equal! %+v != %+v", s, unmarshalled)
	}
}
