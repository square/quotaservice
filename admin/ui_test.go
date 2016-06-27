package admin

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGetDevelopment(t *testing.T) {
	a := NewMockAdministrable()
	handler := NewUIHandler(a, "public", true)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	err := ioutil.WriteFile("public/tempfile.html", []byte("hello"), 777)

	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove("public/tempfile.html")

	res, err := http.Get(ts.URL + "/admin/tempfile.html")

	if err != nil {
		t.Fatal(err)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if err != nil {
		t.Fatal(err)
	}

	body := string(bytes)

	if body != "hello" {
		t.Fatalf("development did not reload and catch /admin/tempfile:\n%s", body)
	}
}

func TestGet(t *testing.T) {
	handler := NewUIHandler(NewMockAdministrable(), "public", false)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/admin/")

	if err != nil {
		t.Fatal(err)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if err != nil {
		t.Fatal(err)
	}

	body := string(bytes)

	if !strings.HasPrefix(body, "<!DOCTYPE html>") {
		t.Fatalf("Received invalid html from /admin/:\n%s", body)
	}
}
