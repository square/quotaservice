// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/maniksurtani/quotaservice/config"
	pb "github.com/maniksurtani/quotaservice/protos/config"
)

func TestNamespacesGetEmpty(t *testing.T) {
	a := NewMockAdministrable()

	configResponse := &pb.ServiceConfig{}
	doNamespacesRequest(t, a, configResponse, "GET", "/api/", "")
}

func TestNamespacesGetNonExistent(t *testing.T) {
	jsonResponse := make(map[string]string)
	doNamespacesRequest(t, NewMockAdministrable(), &jsonResponse, "GET", "/api/test", "")

	if jsonResponse["description"] != "Unable to locate namespace test" {
		t.Errorf("Received \"%s\" from %+v instead of \"Unable to locate namespace test\"",
			jsonResponse["description"], jsonResponse)
	}
}

func TestNamespacesGet(t *testing.T) {
	a := NewMockAdministrable()

	testNamespace := config.NewDefaultNamespaceConfig("test")
	a.Configs().Namespaces["test"] = testNamespace

	configResponse := &pb.NamespaceConfig{}
	doNamespacesRequest(t, a, configResponse, "GET", "/api/test", "")

	if testNamespace.Name != configResponse.Name {
		t.Errorf("Received \"%+v\" but was expecting \"%+v\"", configResponse, testNamespace)
	}
}

func TestNamespacesPost(t *testing.T) {
	jsonResponse := make(map[string]string)
	doNamespacesRequest(t, NewMockAdministrable(), &jsonResponse, "POST", "/api/test", "")

	if len(jsonResponse) != 0 {
		t.Errorf("Received non-empty response \"%+v\"", jsonResponse)
	}
}

func TestNamespacesPostError(t *testing.T) {
	jsonResponse := make(map[string]string)
	doNamespacesRequest(t, NewMockErrorAdministrable(), &jsonResponse, "POST", "/api/test", "")

	if jsonResponse["description"] != "AddNamespace" {
		t.Errorf("Received \"%s\" from %+v instead of AddNamespace", jsonResponse["description"], jsonResponse)
	}
}

func TestNamespacesPut(t *testing.T) {
	jsonResponse := make(map[string]string)
	doNamespacesRequest(t, NewMockAdministrable(), &jsonResponse, "PUT", "/api/test", "")

	if len(jsonResponse) != 0 {
		t.Errorf("Received non-empty response \"%+v\"", jsonResponse)
	}
}

func TestNamespacesPutError(t *testing.T) {
	jsonResponse := make(map[string]string)
	doNamespacesRequest(t, NewMockErrorAdministrable(), &jsonResponse, "PUT", "/api/test", "")

	if jsonResponse["description"] != "UpdateNamespace" {
		t.Errorf("Received \"%s\" from %+v instead of UpdateNamespace", jsonResponse["description"], jsonResponse)
	}
}

func TestNamespacesDeleteError(t *testing.T) {
	jsonResponse := make(map[string]string)
	doNamespacesRequest(t, NewMockErrorAdministrable(), &jsonResponse, "DELETE", "/api/test", "")

	if jsonResponse["description"] != "DeleteNamespace" {
		t.Errorf("Received \"%s\" from %+v instead of DeleteNamespace", jsonResponse["description"], jsonResponse)
	}
}

func TestNamespacesDelete(t *testing.T) {
	jsonResponse := make(map[string]string)
	doNamespacesRequest(t, NewMockAdministrable(), &jsonResponse, "DELETE", "/api/test", "")

	if len(jsonResponse) != 0 {
		t.Errorf("Received non-empty response \"%+v\"", jsonResponse)
	}
}

func doNamespacesRequest(t *testing.T, a Administrable, object interface{}, method, path, body string) {
	apiHandler := NewNamespacesAPIHandler(a)
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
