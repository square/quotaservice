// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConfigsGet(t *testing.T) {
	a := NewMockAdministrable()

	configResponse := &configsResponse{}
	doConfigsRequest(t, a, configResponse, "GET", "/api/configs", "")

	if len(configResponse.Configs) != 1 {
		t.Errorf("Received invalid configs response: %+v", configResponse)
	}
}

func TestConfigsGetError(t *testing.T) {
	a := NewMockErrorAdministrable()

	jsonResponse := make(map[string]string)
	doConfigsRequest(t, a, &jsonResponse, "GET", "/api/configs", "")

	if jsonResponse["description"] != "Error reading configs HistoricalConfigs" {
		t.Errorf("Received \"%s\" from %+v instead of \"Error reading configs HistoricalConfigs\"",
			jsonResponse["description"], jsonResponse)
	}
}

func TestConfigsPut(t *testing.T) {
	a := NewMockAdministrable()

	jsonResponse := make(map[string]string)
	doConfigsRequest(t, a, &jsonResponse, "PUT", "/api/configs", "")

	if jsonResponse["description"] != "Unknown method PUT" {
		t.Errorf("Received \"%s\" from %+v instead of \"Unknown method PUT\"",
			jsonResponse["description"], jsonResponse)
	}
}

func doConfigsRequest(t *testing.T, a Administrable, object interface{}, method, path, body string) {
	apiHandler := NewConfigsAPIHandler(a)
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
