// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package config

import (
	"testing"

	pb "github.com/square/quotaservice/protos/config"
)

func defaultConfig() *pb.ServiceConfig {
	cfg := NewDefaultServiceConfig()

	ns := NewDefaultNamespaceConfig("testNamespace")
	cfg.Namespaces["testNamespace"] = ns

	bt := NewDefaultBucketConfig("testBucket")
	ns.Buckets["testBucket"] = bt

	return cfg
}

func TestGlobalCreateBucket(t *testing.T) {
	cfg := defaultConfig()
	bucket := NewDefaultBucketConfig("newBucket")

	err := CreateBucket(cfg, GlobalNamespace, bucket)

	if err != nil {
		t.Fatalf("CreateBucket errored: %+v", err)
	}

	if cfg.GlobalDefaultBucket != bucket {
		t.Error("newBucket was not created and added to config")
	}

	err = CreateBucket(cfg, GlobalNamespace, bucket)

	if err == nil {
		t.Error("CreateBucket was supposed to error on duplicate DefaultBucket")
	}
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

	bucket = NewDefaultBucketConfig(DefaultBucketName)
	err = CreateBucket(cfg, "testNamespace", bucket)

	if err != nil {
		t.Fatalf("CreateBucket errored: %+v", err)
	}

	if cfg.Namespaces["testNamespace"].DefaultBucket != bucket {
		t.Error("CreateBucket should have set DefaultBucket on namespace")
	}

	cfg.Namespaces["testNamespace"].DefaultBucket = bucket
	err = CreateBucket(cfg, "testNamespace", bucket)

	if err == nil {
		t.Fatalf("CreateBucket should have errored on duplicate DefaultBucket")
	}

	bucket = NewDefaultBucketConfig(DynamicBucketTemplateName)
	err = CreateBucket(cfg, "testNamespace", bucket)

	if err != nil {
		t.Fatalf("CreateBucket errored: %+v", err)
	}

	if cfg.Namespaces["testNamespace"].DynamicBucketTemplate != bucket {
		t.Error("CreateBucket should have set DynamicBucketTemplate on namespace")
	}

	cfg.Namespaces["testNamespace"].DynamicBucketTemplate = bucket
	err = CreateBucket(cfg, "testNamespace", bucket)

	if err == nil {
		t.Fatalf("CreateBucket should have errored on duplicate DynamicBucketTemplate")
	}
}

func TestGlobalUpdateBucket(t *testing.T) {
	cfg := defaultConfig()
	bucket := NewDefaultBucketConfig("newBucket")

	err := UpdateBucket(cfg, GlobalNamespace, bucket)

	if err != nil {
		t.Fatalf("CreateBucket errored: %+v", err)
	}

	if cfg.GlobalDefaultBucket != bucket {
		t.Error("newBucket was not created and added to config")
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

	bucket = NewDefaultBucketConfig(DefaultBucketName)
	err = UpdateBucket(cfg, "testNamespace", bucket)

	if err != nil {
		t.Fatalf("UpdateBucket errored: %+v", err)
	}

	if cfg.Namespaces["testNamespace"].DefaultBucket != bucket {
		t.Error("CreateBucket should have set DefaultBucket on namespace")
	}

	bucket = NewDefaultBucketConfig(DynamicBucketTemplateName)
	err = UpdateBucket(cfg, "testNamespace", bucket)

	if err != nil {
		t.Fatalf("UpdateBucket errored: %+v", err)
	}

	if cfg.Namespaces["testNamespace"].DynamicBucketTemplate != bucket {
		t.Error("UpdateBucket should have set DynamicBucketTemplate on namespace")
	}
}

func TestGlobalDeleteBucket(t *testing.T) {
	cfg := defaultConfig()
	cfg.GlobalDefaultBucket = NewDefaultBucketConfig("newBucket")

	err := DeleteBucket(cfg, GlobalNamespace, "")

	if err != nil {
		t.Fatalf("CreateBucket errored: %+v", err)
	}

	if cfg.GlobalDefaultBucket != nil {
		t.Error("global default bucket was not removed")
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

	bucket := NewDefaultBucketConfig(DefaultBucketName)
	cfg.Namespaces["testNamespace"].DefaultBucket = bucket
	err = DeleteBucket(cfg, "testNamespace", DefaultBucketName)

	if err != nil {
		t.Fatalf("DeleteBucket errored: %+v", err)
	}

	if cfg.Namespaces["testNamespace"].DefaultBucket != nil {
		t.Error("DeleteBucket should have removed DefaultBucket")
	}

	bucket = NewDefaultBucketConfig(DynamicBucketTemplateName)
	cfg.Namespaces["testNamespace"].DynamicBucketTemplate = bucket
	err = DeleteBucket(cfg, "testNamespace", DynamicBucketTemplateName)

	if err != nil {
		t.Fatalf("DeleteBucket errored: %+v", err)
	}

	if cfg.Namespaces["testNamespace"].DynamicBucketTemplate != nil {
		t.Error("DeleteBucket should have removed DynamicBucketTemplate")
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
