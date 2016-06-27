package rest

import (
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/maniksurtani/quotaservice"
	"github.com/maniksurtani/quotaservice/admin"
	"github.com/maniksurtani/quotaservice/config"

	pb "github.com/maniksurtani/quotaservice/protos/config"
)

func TestReadConfigs(t *testing.T) {
	s, c := startService(true, namespaceConfig("ns", true, bucketConfig("b1"), bucketConfig("b2")))
	defer s.Stop()
	// Start an HTTP server
	mux := http.NewServeMux()
	p, e := config.NewDiskConfigPersister("/tmp/qscfgs.dat")
	assertNoError(t, e)
	s.ServeAdminConsole(mux, "", p, false)
	go http.ListenAndServe("127.0.0.1:11111", mux)
	waitForAdminServer()

	rsp, e := http.Get("http://127.0.0.1:11111/api/")
	assertNoError(t, e)
	defer rsp.Body.Close()
	contents, e := ioutil.ReadAll(rsp.Body)
	assertNoError(t, e)

	fromJSON, e := config.FromJSON(contents)
	assertNoError(t, e)

	if !reflect.DeepEqual(fromJSON, c) {
		t.Fatalf("Configs read from API aren't the same as what was expected. Expected: %+v API sent: %+v", c, fromJSON)
	}
}

func TestReadNamespaceConfigs(t *testing.T) {
	s, c := startService(true, namespaceConfig("ns", true, bucketConfig("b1"), bucketConfig("b2")))
	defer s.Stop()
	// Start an HTTP server
	mux := http.NewServeMux()
	p, e := config.NewDiskConfigPersister("/tmp/qscfgs.dat")
	assertNoError(t, e)
	s.ServeAdminConsole(mux, "", p, false)
	go http.ListenAndServe("127.0.0.1:11111", mux)
	waitForAdminServer()

	rsp, e := http.Get("http://127.0.0.1:11111/api/ns")
	assertNoError(t, e)
	defer rsp.Body.Close()
	contents, e := ioutil.ReadAll(rsp.Body)
	assertNoError(t, e)

	fromJSON, e := config.NamespaceFromJSON(contents)
	assertNoError(t, e)

	if !reflect.DeepEqual(fromJSON, c.Namespaces["ns"]) {
		t.Fatalf("Configs read from API aren't the same as what was expected. Expected: %+v API sent: %+v", c.Namespaces["ns"], fromJSON)
	}
}

func TestAddGlobalDefault(t *testing.T) {
	s, _ := startService(false)
	defer s.Stop()

	assertDefaultBucketDoesNotExist(t, s)

	b := config.NewDefaultBucketConfig(config.DefaultBucketName)
	e := s.(admin.Administrable).AddBucket(config.GlobalNamespace, b)
	assertNoError(t, e)

	assertDefaultBucketExists(t, s)

	// Now try and add a bucket config again - should error.
	e = s.(admin.Administrable).AddBucket(config.GlobalNamespace, b)
	assertError(t, e)
}

func TestRemoveGlobalDefault(t *testing.T) {
	s, _ := startService(true)
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
	s, _ := startService(true)
	defer s.Stop()

	assertDefaultBucketExists(t, s)

	b := config.NewDefaultBucketConfig(config.DefaultBucketName)
	b.MaxTokensPerRequest = 2
	e := s.(admin.Administrable).UpdateBucket(config.GlobalNamespace, b)
	assertNoError(t, e)

	// Now check that we hit max tokens limits.
	_, e = s.(quotaservice.QuotaService).Allow("doesn't exist", "doesn't exist", 5, 0)
	assertError(t, e)
	if e.(quotaservice.QuotaServiceError).Reason != quotaservice.ER_TOO_MANY_TOKENS_REQUESTED {
		t.Fatal("Wrong error: ", e)
	}

	b.MaxTokensPerRequest = 10
	e = s.(admin.Administrable).UpdateBucket(config.GlobalNamespace, b)
	assertNoError(t, e)

	// Now check again
	_, e = s.(quotaservice.QuotaService).Allow("doesn't exist", "doesn't exist", 5, 0)
	assertNoError(t, e)
}

func TestAddNamespace(t *testing.T) {
	s, _ := startService(false)
	defer s.Stop()

	assertBucketDoesNotExist(t, s, "ns", "b")

	n := namespaceConfig("ns", true)

	e := s.(admin.Administrable).AddNamespace(n)
	assertNoError(t, e)

	assertBucketExists(t, s, "ns", "b")
	assertBucketExists(t, s, "ns", "bb")
	assertBucketExists(t, s, "ns", "bbb")

	e = s.(admin.Administrable).AddNamespace(n)
	assertError(t, e)

	assertBucketExists(t, s, "ns", "bbbb")
	assertBucketExists(t, s, "ns", "bbbbb")
	assertBucketExists(t, s, "ns", "bbbbbb")
}

func TestRemoveNamespace(t *testing.T) {
	s, _ := startService(false, namespaceConfig("ns", true))
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
	n := namespaceConfig("ns", true)
	s, _ := startService(false, n)
	defer s.Stop()

	// Allows dynamic buckets.
	assertBucketExists(t, s, "ns", "b")
	assertBucketExists(t, s, "ns", "bb")
	assertBucketExists(t, s, "ns", "bbb")

	// change config to not allow dynamic buckets
	n.DynamicBucketTemplate = nil
	e := s.(admin.Administrable).UpdateNamespace(n)
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
	b := bucketConfig("b")
	s, _ := startService(false, namespaceConfig("ns", false, b))
	defer s.Stop()

	// Doesn't allow dynamic buckets.
	assertBucketExists(t, s, "ns", "b")
	assertBucketDoesNotExist(t, s, "ns", "b1")

	// Add bucket
	b.Name = "b1"
	e := s.(admin.Administrable).AddBucket("ns", b)
	assertNoError(t, e)

	// Existing buckets should still be there
	assertBucketExists(t, s, "ns", "b")
	assertBucketExists(t, s, "ns", "b1")
	assertBucketDoesNotExist(t, s, "ns", "b2")

	// Already exists
	e = s.(admin.Administrable).AddBucket("ns", b)
	assertError(t, e)
}

func TestRemoveBucket(t *testing.T) {
	s, _ := startService(false, namespaceConfig("ns", false, bucketConfig("b")))
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
	b := bucketConfig("b")
	s, _ := startService(false, namespaceConfig("ns", false, b))
	defer s.Stop()

	assertBucketExists(t, s, "ns", "b")

	// Now check that we hit max tokens limits.
	_, e := s.(quotaservice.QuotaService).Allow("ns", "b", 5, 0)
	assertError(t, e)
	if e.(quotaservice.QuotaServiceError).Reason != quotaservice.ER_TOO_MANY_TOKENS_REQUESTED {
		t.Fatal("Wrong error: ", e)
	}

	// Update bucket
	b.MaxTokensPerRequest = 10
	e = s.(admin.Administrable).UpdateBucket("ns", b)
	assertNoError(t, e)

	_, e = s.(quotaservice.QuotaService).Allow("ns", "b", 5, 0)
	assertNoError(t, e)
}

func namespaceConfig(n string, dynamic bool, b ...*pb.BucketConfig) *pb.NamespaceConfig {
	ns := config.NewDefaultNamespaceConfig(n)
	for _, bc := range b {
		config.AddBucket(ns, bc)
	}

	if dynamic {
		config.SetDynamicBucketTemplate(ns, config.NewDefaultBucketConfig(""))
	}

	return ns
}

func bucketConfig(n string) *pb.BucketConfig {
	b := config.NewDefaultBucketConfig(n)
	b.MaxTokensPerRequest = 2
	return b
}

func startService(withDefault bool, ns ...*pb.NamespaceConfig) (quotaservice.Server, *pb.ServiceConfig) {
	c := config.NewDefaultServiceConfig()
	if withDefault {
		c.GlobalDefaultBucket = config.NewDefaultBucketConfig(config.DefaultBucketName)
	}
	for _, n := range ns {
		config.AddNamespace(c, n)
	}
	s := quotaservice.New(c, &quotaservice.MockBucketFactory{}, &quotaservice.MockEndpoint{})
	s.Start()
	return s, c
}

func assertDefaultBucketExists(t *testing.T, s quotaservice.Server) {
	assertBucketExists(t, s, "doesn't exist", "doesn't exist")
}

func assertBucketExists(t *testing.T, s quotaservice.Server, nsName, bName string) {
	// Demonstrate that we now do have a default bucket.
	w, e := s.(quotaservice.QuotaService).Allow(nsName, bName, 1, 0)
	assertNoError(t, e)

	if w != 0 {
		t.Fatal("Expecting wait time of 0")
	}
}

func assertDefaultBucketDoesNotExist(t *testing.T, s quotaservice.Server) {
	assertBucketDoesNotExist(t, s, "doesn't exist", "doesn't exist")
}

func assertBucketDoesNotExist(t *testing.T, s quotaservice.Server, nsName, bName string) {
	// Demonstrate that there is no default bucket first
	_, e := s.(quotaservice.QuotaService).Allow(nsName, bName, 1, 0)
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

func waitForAdminServer() {
	maxWait := 2 * time.Minute
	deadline := time.Now().Add(maxWait)

	// Yuck.
	for time.Now().Before(deadline) {
		_, e := http.Get("http://127.0.0.1:11111/")
		if e == nil {
			return
		}
	}

	panic("Waited for 2 minutes, admin server did not come up")
}
