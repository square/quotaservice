// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/square/quotaservice/config"
	pb "github.com/square/quotaservice/protos/config"
)

func TestBucketsGetNamespaceNotFound(t *testing.T) {
	a := NewMockAdministrable()

	jsonResponse := make(map[string]string)
	doBucketsRequest(t, a, &jsonResponse, "GET", "/api/ns/bucket", "")

	if jsonResponse["description"] != "Unable to locate namespace ns" {
		t.Errorf("Received \"%s\" from %+v instead of not found", jsonResponse["description"], jsonResponse)
	}
}

func TestBucketsGetBucketNotFound(t *testing.T) {
	a := NewMockAdministrable()
	testNamespace := config.NewDefaultNamespaceConfig("test")
	a.Configs().Namespaces["test"] = testNamespace

	jsonResponse := make(map[string]string)
	doBucketsRequest(t, a, &jsonResponse, "GET", "/api/test/bucket", "")

	if jsonResponse["description"] != "Unable to locate bucket bucket in namespace test" {
		t.Errorf("Received \"%s\" from %+v instead of not found", jsonResponse["description"], jsonResponse)
	}
}

func TestBucketsGet(t *testing.T) {
	a := NewMockAdministrable()

	testNamespace := config.NewDefaultNamespaceConfig("test")
	a.Configs().Namespaces["test"] = testNamespace
	bucket := config.NewDefaultBucketConfig("bucket")
	testNamespace.Buckets["bucket"] = bucket

	configResponse := &pb.BucketConfig{}
	doBucketsRequest(t, a, configResponse, "GET", "/api/test/bucket", "")

	if *bucket != *configResponse {
		t.Errorf("Received \"%+v\" but was expecting \"%+v\"", configResponse, bucket)
	}
}

func TestBucketsPost(t *testing.T) {
	jsonResponse := make(map[string]string)
	doBucketsRequest(t, NewMockAdministrable(), &jsonResponse, "POST", "/api/test/newbucket", "")

	if len(jsonResponse) != 0 {
		t.Errorf("Received non-empty response \"%+v\"", jsonResponse)
	}
}

func TestBucketsPostError(t *testing.T) {
	jsonResponse := make(map[string]string)
	doBucketsRequest(t, NewMockErrorAdministrable(), &jsonResponse, "POST", "/api/ns/newbucket", "")

	if jsonResponse["description"] != "AddBucket" {
		t.Errorf("Received \"%s\" from %+v instead of AddBucket", jsonResponse["description"], jsonResponse)
	}
}

func TestBucketsPut(t *testing.T) {
	jsonResponse := make(map[string]string)
	doBucketsRequest(t, NewMockAdministrable(), &jsonResponse, "PUT", "/api/test/newbucket", "")

	if len(jsonResponse) != 0 {
		t.Errorf("Received non-empty response \"%+v\"", jsonResponse)
	}
}

func TestBucketsPutError(t *testing.T) {
	jsonResponse := make(map[string]string)
	doBucketsRequest(t, NewMockErrorAdministrable(), &jsonResponse, "PUT", "/api/ns/newbucket", "")

	if jsonResponse["description"] != "UpdateBucket" {
		t.Errorf("Received \"%s\" from %+v instead of UpdateBucket", jsonResponse["description"], jsonResponse)
	}
}

func TestBucketsDeleteError(t *testing.T) {
	jsonResponse := make(map[string]string)
	doBucketsRequest(t, NewMockErrorAdministrable(), &jsonResponse, "DELETE", "/api/ns/bucket", "")

	if jsonResponse["description"] != "DeleteBucket" {
		t.Errorf("Received \"%s\" from %+v instead of DeleteBucket", jsonResponse["description"], jsonResponse)
	}
}

func TestBucketsDelete(t *testing.T) {
	jsonResponse := make(map[string]string)
	doBucketsRequest(t, NewMockAdministrable(), &jsonResponse, "DELETE", "/api/ns/bucket", "")

	if len(jsonResponse) != 0 {
		t.Errorf("Received non-empty response \"%+v\"", jsonResponse)
	}
}

func doBucketsRequest(t *testing.T, a Administrable, object interface{}, method, path, body string) {
	t.Helper()

	apiHandler := newBucketsAPIHandler(a)
	ts := httptest.NewServer(apiHandler)
	defer ts.Close()

	client := &http.Client{}
	request, err := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	res, err := client.Do(request)

	if err != nil {
		t.Fatal(err)
	}

	err = unmarshalJSON(res.Body, &object)
	if err != nil {
		t.Fatal(err)
	}
}
