// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package config

import (
	"testing"

	pb "github.com/maniksurtani/quotaservice/protos/config"
)

func defaultConfig() *pb.ServiceConfig {
	cfg := NewDefaultServiceConfig()

	ns := NewDefaultNamespaceConfig("testNamespace")
	cfg.Namespaces["testNamespace"] = ns

	bt := NewDefaultBucketConfig("testBucket")
	ns.Buckets["testBucket"] = bt

	return cfg
}

func TestCreateBucket(t *testing.T) {
	cfg := defaultConfig()
	bucket := NewDefaultBucketConfig("newBucket")

	err := CreateBucket(cfg, "nilNamespace", bucket)

	if err == nil {
		t.Error("CreateBucket was supposed to error on nonexistent namespace")
	}

	err = CreateBucket(cfg, "testNamespace", bucket)

	if err != nil {
		t.Fatalf("CreateBucket errored: %+v", err)
	}

	if cfg.Namespaces["testNamespace"].Buckets["newBucket"] != bucket {
		t.Error("newBucket was not created and added to config")
	}

	bucket = NewDefaultBucketConfig("testBucket")
	err = CreateBucket(cfg, "testNamespace", bucket)

	if err == nil {
		t.Error("CreateBucket was supposed to error on dup bucket")
	}
}

func TestUpdateBucket(t *testing.T) {
	cfg := defaultConfig()
	bucket := NewDefaultBucketConfig("testBucket")

	err := UpdateBucket(cfg, "nilNamespace", bucket)

	if err == nil {
		t.Error("UpdateBucket was supposed to error on nonexistent namespace")
	}

	err = UpdateBucket(cfg, "testNamespace", bucket)

	if err != nil {
		t.Fatalf("UpdateBucket errored: %+v", err)
	}

	if cfg.Namespaces["testNamespace"].Buckets["testBucket"] != bucket {
		t.Error("testBucket was not updated")
	}
}

func TestDeleteBucket(t *testing.T) {
	cfg := defaultConfig()

	err := DeleteBucket(cfg, "nilNamespace", "testBucket")

	if err == nil {
		t.Error("UpdateBucket was supposed to error on nonexistent namespace")
	}

	err = DeleteBucket(cfg, "testNamespace", "testBucket")

	if err != nil {
		t.Fatalf("UpdateBucket errored: %+v", err)
	}

	if cfg.Namespaces["testNamespace"].Buckets["testBucket"] != nil {
		t.Error("testBucket was not removed")
	}
}

func TestDeleteNamespace(t *testing.T) {
	cfg := defaultConfig()

	err := DeleteNamespace(cfg, "nonexistent")

	if err == nil {
		t.Error("DeleteNamespace did not error on nonexistent namespace")
	}

	err = DeleteNamespace(cfg, "testNamespace")

	if err != nil {
		t.Fatalf("DeleteNamespace errored: %+v", err)
	}

	if cfg.Namespaces["testNamespace"] != nil {
		t.Error("DeleteNamespace did not remove testNamespace")
	}
}

func TestCreateNamespace(t *testing.T) {
	cfg := defaultConfig()
	ns := NewDefaultNamespaceConfig("newNamespace")

	err := CreateNamespace(cfg, ns)

	if err != nil {
		t.Fatalf("CreateNamespace errored: %+v", err)
	}

	if cfg.Namespaces["newNamespace"] != ns {
		t.Error("CreateNamespace did not add newNamespace properly")
	}

	ns = NewDefaultNamespaceConfig("testNamespace")
	err = CreateNamespace(cfg, ns)

	if err == nil {
		t.Error("CreateNamespace did not error on duplicate namespace")
	}
}

func TestUpdateNamespace(t *testing.T) {
	cfg := defaultConfig()
	ns := NewDefaultNamespaceConfig("newNamespace")

	err := UpdateNamespace(cfg, ns)

	if err != nil {
		t.Fatalf("UpdateNamespace errored: %+v", err)
	}

	if cfg.Namespaces["newNamespace"] != ns {
		t.Error("UpdateNamespace did not add newNamespace properly")
	}

	ns = NewDefaultNamespaceConfig("testNamespace")
	err = UpdateNamespace(cfg, ns)

	if err != nil {
		t.Fatalf("UpdateNamespace errored: %+v", err)
	}

	if cfg.Namespaces["testNamespace"] != ns {
		t.Error("UpdateNamespace did not update testNamespace")
	}

	newCfg := &pb.ServiceConfig{}
	err = UpdateNamespace(newCfg, ns)

	if err != nil {
		t.Fatalf("UpdateNamespace errored: %+v", err)
	}

	if newCfg.Namespaces["testNamespace"] != ns {
		t.Error("UpdateNamespace did not update testNamespace")
	}
}
