// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

// Package implements admin UIs and a REST API
package admin

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/maniksurtani/quotaservice/config"
	"github.com/maniksurtani/quotaservice/logging"
	pb "github.com/maniksurtani/quotaservice/protos/config"
	"encoding/json"
	"io"
)

// Administrable defines something that can be administered via this package.
type Administrable interface {
	Configs() *config.ServiceConfig

	DeleteBucket(namespace, name string) error
	AddBucket(namespace string, b *pb.BucketConfig) error
	UpdateBucket(namespace string, b *pb.BucketConfig) error

	DeleteNamespace(namespace string) error
	AddNamespace(n *pb.NamespaceConfig) error
	UpdateNamespace(n *pb.NamespaceConfig) error
}

// ServeAdminConsole serves up an admin console for an Administrable over a http server. assetsDirectory contains
// HTML templates and other UI assets. If empty, no UI will be served, and only REST endpoints under /api/ will be
// served instead.
func ServeAdminConsole(a Administrable, mux *http.ServeMux, assetsDirectory string) {
	logging.Print("Serving admin console.")
	if assetsDirectory != "" {
		files, err := ioutil.ReadDir(assetsDirectory)
		check(err)
		htmlFiles := make([]string, 0)
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".html") {
				htmlFiles = append(htmlFiles, assetsDirectory + "/" + f.Name())
			}
		}
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/admin/", 301)
		})
		mux.Handle("/admin/", &uiHandler{a, reloadTemplates(htmlFiles), htmlFiles})
		mux.Handle("/js/", http.FileServer(http.Dir(assetsDirectory)))
	} else {
		logging.Print("Not serving UI.")
	}
	mux.Handle("/api/", &apiHandler{a})
}

type uiHandler struct {
	a Administrable
	t *template.Template
	h []string
}

func reloadTemplates(files []string) *template.Template {
	return template.Must(template.New("admin").ParseFiles(files...))
}

func (h *uiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO(manik) remove this
	h.t = reloadTemplates(h.h)

	path := r.URL.Path[len("/admin/"):]

	var tpl string

	if path == "" || path == "/" {
		tpl = "index.html"
	} else {
		tpl = path
	}

	err := h.t.ExecuteTemplate(w, tpl, h.a.Configs())
	if err != nil {
		logging.Printf("Caught error %v serving URL %v", err, r.URL.Path)
		http.NotFound(w, r)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type apiHandler struct {
	a Administrable
}

func (a *apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/bucket/") {
		params := strings.TrimPrefix(r.URL.Path, "/api/bucket/")
		namespace, name := extractNamespaceName(params)
		logging.Printf("Request for bucket %v", params)
		switch r.Method {
		case "DELETE":
			a.a.DeleteBucket(namespace, name)
		case "PUT":
			c, e := getBucketConfig(r.Body)
			if e != nil {
				logging.Println("Caught error", e)
				http.Error(w, "500 bad content", http.StatusInternalServerError)
			} else {
				a.a.AddBucket(namespace, c)
			}
		case "POST":
			c, e := getBucketConfig(r.Body)
			if e != nil {
				logging.Println("Caught error", e)
				http.Error(w, "500 bad content", http.StatusInternalServerError)
			} else {
				a.a.UpdateBucket(namespace, c)
			}
		default:
			logging.Printf("Not handling method %v", r.Method)
			http.NotFound(w, r)
		}
	} else if strings.HasPrefix(r.URL.Path, "/api/namespace/") {
		ns := strings.TrimPrefix(r.URL.Path, "/api/namespace/")
		switch r.Method {
		case "DELETE":
			a.a.DeleteNamespace(ns)
		case "PUT":
			c, e := getNamespaceConfig(r.Body)
			if e != nil {
				logging.Println("Caught error", e)
				http.Error(w, "500 bad content", http.StatusInternalServerError)
			} else {
				a.a.AddNamespace(c)
			}
		case "POST":
			c, e := getNamespaceConfig(r.Body)
			if e != nil {
				logging.Println("Caught error", e)
				http.Error(w, "500 bad content", http.StatusInternalServerError)
			} else {
				a.a.UpdateNamespace(c)
			}
		default:
			logging.Printf("Not handling method %v", r.Method)
			http.NotFound(w, r)
		}
	} else {
		logging.Printf("Not handling path %v", r.URL.Path)
		http.NotFound(w, r)
	}
}

func extractNamespaceName(params string) (namespace, name string) {
	// params should be in the format xyz/abc. We just split on '/'
	parts := strings.Split(params, "/")
	if len(parts) < 2 {
		panic("Params '" + params + "' doesn't contain a '/'")
	}
	return parts[0], parts[1]
}

func getBucketConfig(r io.Reader) (*pb.BucketConfig, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	c := &pb.BucketConfig{}
	json.Unmarshal(bytes, c)
	return c, nil
}

func getNamespaceConfig(r io.Reader) (*pb.NamespaceConfig, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	c := &pb.NamespaceConfig{}
	json.Unmarshal(bytes, c)
	return c, nil
}
