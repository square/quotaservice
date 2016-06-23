// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package config

import (
	pbconfig "github.com/maniksurtani/quotaservice/protos/config"
	"reflect"
	"testing"
)

func TestPersistence(t *testing.T) {
	persister, e := NewDiskConfigPersister("/tmp/qs_test_persistence")
	checkError(t, e)

	select {
	case <-persister.ConfigChangedWatcher():
		t.Fatal("Config channel should be empty!")
	default:
		// This is good.
	}

	s := &pbconfig.ServiceConfig{
		GlobalDefaultBucket: &pbconfig.BucketConfig{Size: 300, FillRate: 400, WaitTimeoutMillis: 123456},
		Namespaces:          make(map[string]*pbconfig.NamespaceConfig, 1),
		Version:             92}

	nc := &pbconfig.NamespaceConfig{
		Name:              "xyz",
		MaxDynamicBuckets: 123}

	SetDynamicBucketTemplate(nc, &pbconfig.BucketConfig{})
	AddNamespace(s, nc)

	// Store s.
	r, e := Marshal(s)
	checkError(t, e)
	e = persister.PersistAndNotify(r)
	checkError(t, e)

	// Test notification
	select {
	case <-persister.ConfigChangedWatcher():
	// This is good.
	default:
		t.Fatal("Config channel should not be empty!")
	}

	r, e = persister.ReadPersistedConfig()
	checkError(t, e)
	unmarshalled, e := Unmarshal(r)
	checkError(t, e)

	if !reflect.DeepEqual(s, unmarshalled) {
		t.Fatalf("Configs should be equal! %+v != %+v", s, unmarshalled)
	}
}

func checkError(t *testing.T, e error) {
	if e != nil {
		t.Fatal("Not expecting error ", e)
	}
}
