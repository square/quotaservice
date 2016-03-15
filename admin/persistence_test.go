// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package admin

import (
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
	pb "github.com/maniksurtani/quotaservice/protos/config"
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

	s := &pb.ServiceConfig{
		GlobalDefaultBucket: &pb.BucketConfig{Size: proto.Int64(300), FillRate: proto.Int64(400), WaitTimeoutMillis: proto.Int64(123456)},
		Namespaces:          make([]*pb.NamespaceConfig, 1),
		Version:             proto.Int(92)}

	s.Namespaces[0] = &pb.NamespaceConfig{Name: proto.String("xyz"), MaxDynamicBuckets: proto.Int(123), DynamicBucketTemplate: &pb.BucketConfig{}}

	// Store s.
	e = persister.PersistAndNotify(s)
	checkError(t, e)

	// Test notification
	select {
	case <-persister.ConfigChangedWatcher():
		// This is good.
	default:
		t.Fatal("Config channel should not be empty!")
	}

	var newCfg *pb.ServiceConfig
	newCfg, e = persister.ReadPersistedConfig()
	checkError(t, e)

	if !reflect.DeepEqual(newCfg, s) {
		t.Fatal("Configs should be equal!")
	}
}

func checkError(t *testing.T, e error) {
	if e != nil {
		t.Fatal("Not expecting error ", e)
	}
}
