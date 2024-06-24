// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/square/quotaservice/config"
	"github.com/square/quotaservice/test/helpers"
)

func TestNamespacesPostWithVersion(t *testing.T) {
	ts := establishTestServer(3)
	defer ts.Close()
	jsonResponse, _ := executeRequestForVersioningTest(ts, true, http.MethodPost, "3", t)

	if jsonResponse["error"] != "" {
		t.Errorf("POST request with correct version header should succeed: %+v", jsonResponse)
	}
}

func TestNamespacesPostNoVersion(t *testing.T) {
	ts := establishTestServer(10)
	defer ts.Close()
	jsonResponse, _ := executeRequestForVersioningTest(ts, false, http.MethodPost, "", t)

	if jsonResponse["error"] != http.StatusText(http.StatusBadRequest) {
		t.Errorf("Expected 400 Bad Request, but received \"%+v\"", jsonResponse)
	}
}

func TestNamespacesPostIncorrectVersion(t *testing.T) {
	ts := establishTestServer(10)
	defer ts.Close()
	jsonResponse, _ := executeRequestForVersioningTest(ts, true, http.MethodPost, "3", t)

	if jsonResponse["error"] != http.StatusText(http.StatusConflict) {
		t.Errorf("Expected 409 Conflict, but received \"%+v\"", jsonResponse)
	}
}

func TestNamespacesPostBadVersion(t *testing.T) {
	ts := establishTestServer(10)
	defer ts.Close()
	jsonResponse, _ := executeRequestForVersioningTest(ts, true, http.MethodPost, "ABC", t)

	if jsonResponse["error"] != http.StatusText(http.StatusBadRequest) {
		t.Errorf("Expected 400 Bad Request, but received \"%+v\"", jsonResponse)
	}
}

func TestNamespacesGetWithVersion(t *testing.T) {
	currentVersion := 3
	ts := establishTestServer(int32(currentVersion))
	defer ts.Close()
	jsonResponse, resVersion := executeRequestForVersioningTest(ts, true, http.MethodGet, "3", t)

	if resVersion != fmt.Sprintf("%d", currentVersion) {
		t.Errorf("Expected version %d, but received %s", currentVersion, resVersion)
	}

	if jsonResponse["error"] != "" {
		t.Errorf("GET request with correct version header should succeed: %+v", jsonResponse)
	}
}

func TestNamespacesGetNoVersion(t *testing.T) {
	ts := establishTestServer(10)
	defer ts.Close()
	jsonResponse, _ := executeRequestForVersioningTest(ts, false, http.MethodGet, "", t)

	if jsonResponse["error"] != "" {
		t.Errorf("GET request without version header should succeed: %+v", jsonResponse)
	}
}

func TestNamespacesGetIncorrectVersion(t *testing.T) {
	ts := establishTestServer(10)
	defer ts.Close()
	jsonResponse, _ := executeRequestForVersioningTest(ts, true, http.MethodGet, "3", t)

	if jsonResponse["error"] != "" {
		t.Errorf("GET request with incorrect version header should succeed, as it's ignored: %+v", jsonResponse)
	}
}

func TestNamespacesGetBadVersion(t *testing.T) {
	ts := establishTestServer(10)
	defer ts.Close()
	jsonResponse, _ := executeRequestForVersioningTest(ts, true, http.MethodGet, "ABC", t)

	if jsonResponse["error"] != "" {
		t.Errorf("GET request with bad version header should succeed, as it's ignored: %+v", jsonResponse)
	}
}

func TestUnmarshalBucketConfig(t *testing.T) {
	c := config.NewDefaultBucketConfig("Blah 123")
	c.FillRate = 12345
	c.MaxDebtMillis = 54321
	c.MaxIdleMillis = 67890
	c.MaxTokensPerRequest = 9876
	c.Size = 50000

	b, e := json.Marshal(c)
	if e != nil {
		t.Fatal("Unable to JSONify proto", e)
	}

	reRead, err := getBucketConfig(bytes.NewReader(b))
	if err != nil {
		t.Fatal("Unable to unmarshal JSON", err)
	}
	if !reflect.DeepEqual(c, reRead) {
		t.Fatalf("Two representations aren't equal: %+v != %+v", c, reRead)
	}
}

func TestUnmarshalNamespaceConfig(t *testing.T) {
	n := config.NewDefaultNamespaceConfig("Blah Namespace 123")
	n.MaxDynamicBuckets = 8000
	config.SetDynamicBucketTemplate(n, config.NewDefaultBucketConfig(""))

	c1 := config.NewDefaultBucketConfig("Blah 123")
	c1.FillRate = 12345
	c1.MaxDebtMillis = 54321
	c1.MaxIdleMillis = 67890
	c1.MaxTokensPerRequest = 9876
	c1.Size = 50000

	c2 := config.NewDefaultBucketConfig("Blah 456")
	c2.FillRate = 123450
	c2.MaxDebtMillis = 543210
	c2.MaxIdleMillis = 678900
	c2.MaxTokensPerRequest = 98760
	c2.Size = 5000

	c3 := config.NewDefaultBucketConfig("Blah 789")
	c3.FillRate = 1234500
	c3.MaxDebtMillis = 5432100
	c3.MaxIdleMillis = 6789000
	c3.MaxTokensPerRequest = 987600
	c3.Size = 500

	helpers.CheckError(t, config.AddBucket(n, c1))
	helpers.CheckError(t, config.AddBucket(n, c2))
	helpers.CheckError(t, config.AddBucket(n, c3))

	b, e := json.Marshal(n)
	if e != nil {
		t.Fatal("Unable to JSONify proto", e)
	}

	reRead, err := getNamespaceConfig(bytes.NewReader(b))
	if err != nil {
		t.Fatal("Unable to unmarshal JSON", err)
	}

	if !reflect.DeepEqual(n, reRead) {
		t.Fatalf("Two representations aren't equal: %+v != %+v", n, reRead)
	}
}

func establishTestServer(version int32) *httptest.Server {
	a := NewMockAdministrable()
	a.Configs().Version = version

	apiHandler := apiVersionHandler(a, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			writeJSONOk(w)
		}),
	)

	return httptest.NewServer(apiHandler)
}

func executeRequestForVersioningTest(ts *httptest.Server, versioned bool, method string, version string, t *testing.T) (jsonResponse map[string]string, resVersion string) {
	client := &http.Client{}
	request, err := http.NewRequest(method, ts.URL, strings.NewReader(""))

	if versioned {
		request.Header.Set("Version", version)
	}

	res, err := client.Do(request)

	if err != nil {
		t.Fatal(err)
	}

	err = unmarshalJSON(res.Body, &jsonResponse)

	if err != nil {
		t.Fatal(err)
	}

	return jsonResponse, res.Header.Get("Version")
}
