// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/maniksurtani/quotaservice/config"
	"github.com/maniksurtani/quotaservice/stats"
)

func TestStatsErrors(t *testing.T) {
	a := NewMockAdministrable()
	jsonResponse := make(map[string]string)

	doStatsRequest(t, a, &jsonResponse, "GET", "/api/stats", "")

	if jsonResponse["description"] != "No namespace specified" {
		t.Errorf("Received \"%s\" from %+v instead of \"No namespace specified\"",
			jsonResponse["description"], jsonResponse)
	}

	doStatsRequest(t, a, &jsonResponse, "DELETE", "/api/stats", "")

	if jsonResponse["description"] != "Unknown method DELETE" {
		t.Errorf("Received \"%s\" from %+v instead of \"Unknown method DELETE\"",
			jsonResponse["description"], jsonResponse)
	}

	doStatsRequest(t, a, &jsonResponse, "GET", "/api/stats/unknown", "")

	if jsonResponse["description"] != "Unable to locate namespace unknown" {
		t.Errorf("Received \"%s\" from %+v instead of \"Unable to locate namespace unknown\"",
			jsonResponse["description"], jsonResponse)
	}
}

func TestStatsGet(t *testing.T) {
	a := NewMockAdministrable()

	testNamespace := config.NewDefaultNamespaceConfig("test")
	a.Configs().Namespaces["test"] = testNamespace

	nsResponse := &bucketStats{}
	doStatsRequest(t, a, nsResponse, "GET", "/api/stats/test", "")

	if nsResponse.Ns != "test" || len(nsResponse.Hits) != 0 || len(nsResponse.Misses) != 0 {
		t.Errorf("Received %+v instead of [Ns=test, Hits=[], Misses=[]]", nsResponse)
	}

	a = NewMockErrorAdministrable()
	a.Configs().Namespaces["test"] = testNamespace

	jsonResponse := make(map[string]string)

	doStatsRequest(t, a, &jsonResponse, "GET", "/api/stats/test", "")

	if jsonResponse["description"] != "No stats listener configured" {
		t.Errorf("Received \"%s\" from %+v instead of \"No stats listener configured\"",
			jsonResponse["description"], jsonResponse)
	}
}

func TestStatsGetBucket(t *testing.T) {
	a := NewMockAdministrable()

	testNamespace := config.NewDefaultNamespaceConfig("test")
	a.Configs().Namespaces["test"] = testNamespace
	bucket := config.NewDefaultBucketConfig("bucket")
	testNamespace.Buckets["bucket"] = bucket

	bucketResponse := make(map[string]*stats.BucketScores)
	doStatsRequest(t, a, &bucketResponse, "GET", "/api/stats/test/bucket", "")

	bucketScores, ok := bucketResponse["bucket"]

	if !ok || bucketScores.Hits != 0 || bucketScores.Misses != 0 {
		t.Errorf("Received %+v instead of [bucket=[Hits=0, Misses=0]]", bucketResponse)
	}

	a = NewMockErrorAdministrable()
	a.Configs().Namespaces["test"] = testNamespace
	testNamespace.Buckets["bucket"] = bucket

	jsonResponse := make(map[string]string)
	doStatsRequest(t, a, &jsonResponse, "GET", "/api/stats/test/bucket", "")

	if jsonResponse["description"] != "No stats listener configured" {
		t.Errorf("Received \"%s\" from %+v instead of \"No stats listener configured\"",
			jsonResponse["description"], jsonResponse)
	}
}

func doStatsRequest(t *testing.T, a Administrable, object interface{}, method, path, body string) {
	apiHandler := NewStatsAPIHandler(a)
	ts := httptest.NewServer(apiHandler)
	defer ts.Close()

	client := &http.Client{}
	request, err := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	res, err := client.Do(request)

	if err != nil {
		t.Fatal(err)
	}

	err = unmarshalJSON(res.Body, &object)
	res.Body.Close()

	if err != nil {
		t.Fatal(err)
	}
}
