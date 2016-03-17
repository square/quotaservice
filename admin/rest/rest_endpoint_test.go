package rest

import (
	"testing"
	"github.com/maniksurtani/quotaservice/config"
	"github.com/maniksurtani/quotaservice"
	"github.com/maniksurtani/quotaservice/admin"
	"fmt"
)

func TestAddGlobalDefault(t *testing.T) {
	s := startService(false)
	defer s.Stop()

	assertDefaultBucketDoesNotExist(t, s)

	b := config.NewDefaultBucketConfig()
	e := s.(admin.Administrable).AddBucket(config.GlobalNamespace, b.ToProto())
	assertNoError(t, e)

	assertDefaultBucketExists(t, s)

	// Now try and add a bucket config again - should error.
	e = s.(admin.Administrable).AddBucket(config.GlobalNamespace, b.ToProto())
	assertError(t, e)
}

func TestRemoveGlobalDefault(t *testing.T) {
	s := startService(true)
	defer s.Stop()

	assertDefaultBucketExists(t, s)

	e := s.(admin.Administrable).DeleteBucket(config.GlobalNamespace, config.DefaultBucketName)
	assertNoError(t, e)

	assertDefaultBucketDoesNotExist(t, s)

	// Should be idempotent
	e = s.(admin.Administrable).DeleteBucket(config.GlobalNamespace, config.DefaultBucketName)
	assertNoError(t, e)

	assertDefaultBucketDoesNotExist(t, s)
}

func TestUpdateGlobalDefault(t *testing.T) {
	s := startService(true)
	defer s.Stop()

	assertDefaultBucketExists(t, s)

	b := config.NewDefaultBucketConfig()
	b.MaxTokensPerRequest = 2
	b.Name = config.DefaultBucketName
	e := s.(admin.Administrable).UpdateBucket(config.GlobalNamespace, b.ToProto())
	assertNoError(t, e)

	// Now check that we hit max tokens limits.
	_, _, e = s.(quotaservice.QuotaService).Allow("doesn't exist", "doesn't exist", 5, 0)
	assertError(t, e)
	if e.(quotaservice.QuotaServiceError).Reason != quotaservice.ER_TOO_MANY_TOKENS_REQUESTED {
		t.Fatal("Wrong error: ", e)
	}

	b.MaxTokensPerRequest = 10
	e = s.(admin.Administrable).UpdateBucket(config.GlobalNamespace, b.ToProto())
	assertNoError(t, e)

	// Now check again
	_, _, e = s.(quotaservice.QuotaService).Allow("doesn't exist", "doesn't exist", 5, 0)
	assertNoError(t, e)
}

func TestAddNamespace(t *testing.T) {
	s := startService(false)
	defer s.Stop()

	assertBucketDoesNotExist(t, s, "ns", "b")

	n := config.NewDefaultNamespaceConfig()
	n.Name = "ns"
	n.SetDynamicBucketTemplate(config.NewDefaultBucketConfig())

	e := s.(admin.Administrable).AddNamespace(n.ToProto())
	assertNoError(t, e)

	assertBucketExists(t, s, "ns", "b")
	assertBucketExists(t, s, "ns", "bb")
	assertBucketExists(t, s, "ns", "bbb")

	e = s.(admin.Administrable).AddNamespace(n.ToProto())
	assertError(t, e)

	assertBucketExists(t, s, "ns", "bbbb")
	assertBucketExists(t, s, "ns", "bbbbb")
	assertBucketExists(t, s, "ns", "bbbbbb")
}

func TestRemoveNamespace(t *testing.T) {
	n := config.NewDefaultNamespaceConfig()
	n.Name = "ns"
	n.SetDynamicBucketTemplate(config.NewDefaultBucketConfig())

	s := startService(false, n)
	defer s.Stop()

	assertBucketExists(t, s, "ns", "b")

	e := s.(admin.Administrable).DeleteNamespace("ns")
	assertNoError(t, e)

	assertBucketDoesNotExist(t, s, "ns", "b")

	e = s.(admin.Administrable).DeleteNamespace("ns")
	assertError(t, e)

	assertBucketDoesNotExist(t, s, "ns", "b")
}

func TestUpdateNamespace(t *testing.T) {
	n := config.NewDefaultNamespaceConfig()
	n.Name = "ns"
	n.SetDynamicBucketTemplate(config.NewDefaultBucketConfig())

	s := startService(false, n)
	defer s.Stop()

	// Allows dynamic buckets.
	assertBucketExists(t, s, "ns", "b")
	assertBucketExists(t, s, "ns", "bb")
	assertBucketExists(t, s, "ns", "bbb")

	// change config to not allow dynamic buckets
	n.DynamicBucketTemplate = nil
	e := s.(admin.Administrable).UpdateNamespace(n.ToProto())
	assertNoError(t, e)

	// Existing buckets should have been removed.
	assertBucketDoesNotExist(t, s, "ns", "b")
	assertBucketDoesNotExist(t, s, "ns", "bb")
	assertBucketDoesNotExist(t, s, "ns", "bbb")

	// No new dynamic buckets
	assertBucketDoesNotExist(t, s, "ns", "bbbb")
	assertBucketDoesNotExist(t, s, "ns", "bbbbb")
	assertBucketDoesNotExist(t, s, "ns", "bbbbbb")
}

func TestAddBucket(t *testing.T) {
	n := config.NewDefaultNamespaceConfig()
	n.Name = "ns"
	b := config.NewDefaultBucketConfig()
	b.Name = "b"
	n.AddBucket("b", b)

	s := startService(false, n)
	defer s.Stop()

	// Doesn't allow dynamic buckets.
	assertBucketExists(t, s, "ns", "b")
	assertBucketDoesNotExist(t, s, "ns", "b1")

	// Add bucket
	b.Name = "b1"
	e := s.(admin.Administrable).AddBucket("ns", b.ToProto())
	assertNoError(t, e)

	// Existing buckets should still be there
	assertBucketExists(t, s, "ns", "b")
	assertBucketExists(t, s, "ns", "b1")
	assertBucketDoesNotExist(t, s, "ns", "b2")

	// Already exists
	e = s.(admin.Administrable).AddBucket("ns", b.ToProto())
	assertError(t, e)
}

func TestRemoveBucket(t *testing.T) {
	n := config.NewDefaultNamespaceConfig()
	n.Name = "ns"
	b := config.NewDefaultBucketConfig()
	b.Name = "b"
	n.AddBucket("b", b)

	s := startService(false, n)
	defer s.Stop()

	// Doesn't allow dynamic buckets.
	assertBucketExists(t, s, "ns", "b")
	assertBucketDoesNotExist(t, s, "ns", "b1")

	// Add bucket
	e := s.(admin.Administrable).DeleteBucket("ns", "b")
	assertNoError(t, e)

	// Existing buckets should still be there
	assertBucketDoesNotExist(t, s, "ns", "b")

	// Idempotence
	e = s.(admin.Administrable).DeleteBucket("ns", "b")
	assertNoError(t, e)
}

func TestUpdateBucket(t *testing.T) {
	n := config.NewDefaultNamespaceConfig()
	n.Name = "ns"
	b := config.NewDefaultBucketConfig()
	b.Name = "b"
	b.MaxTokensPerRequest = 2
	n.AddBucket("b", b)

	s := startService(false, n)
	defer s.Stop()

	assertBucketExists(t, s, "ns", "b")

	// Now check that we hit max tokens limits.
	_, _, e := s.(quotaservice.QuotaService).Allow("ns", "b", 5, 0)
	assertError(t, e)
	if e.(quotaservice.QuotaServiceError).Reason != quotaservice.ER_TOO_MANY_TOKENS_REQUESTED {
		t.Fatal("Wrong error: ", e)
	}

	// Update bucket
	b.MaxTokensPerRequest = 10
	e = s.(admin.Administrable).UpdateBucket("ns", b.ToProto())
	assertNoError(t, e)

	_, _, e = s.(quotaservice.QuotaService).Allow("ns", "b", 5, 0)
	assertNoError(t, e)
}

func startService(withDefault bool, ns ...*config.NamespaceConfig) quotaservice.Server {
	c := config.NewDefaultServiceConfig()
	if !withDefault {
		c.GlobalDefaultBucket = nil
	}
	for _, n := range ns {
		fmt.Printf("Adding namespace %+v ", n)
		c.AddNamespace(n.Name, n)
	}
	s := quotaservice.New(c, &quotaservice.MockBucketFactory{}, &quotaservice.MockEndpoint{})
	s.Start()
	return s
}

func assertDefaultBucketExists(t *testing.T, s quotaservice.Server) {
	assertBucketExists(t, s, "doesn't exist", "doesn't exist")
}

func assertBucketExists(t *testing.T, s quotaservice.Server, nsName, bName string) {
	// Demonstrate that we now do have a default bucket.
	g, w, e := s.(quotaservice.QuotaService).Allow(nsName, bName, 1, 0)
	assertNoError(t, e)

	if g != 1 {
		t.Fatal("Expected to be granted 1 token")
	}

	if w != 0 {
		t.Fatal("Expecting wait time of 0")
	}
}

func assertDefaultBucketDoesNotExist(t *testing.T, s quotaservice.Server) {
	assertBucketDoesNotExist(t, s, "doesn't exist", "doesn't exist")
}

func assertBucketDoesNotExist(t *testing.T, s quotaservice.Server, nsName, bName string) {
	// Demonstrate that there is no default bucket first
	_, _, e := s.(quotaservice.QuotaService).Allow(nsName, bName, 1, 0)
	assertError(t, e)
}

func assertNoError(t *testing.T, e error) {
	if e != nil {
		t.Fatal("Not expecting error ", e)
	}
}

func assertError(t *testing.T, e error) {
	if e == nil {
		t.Fatal("Expecting error!")
	}
}
