// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package admin

import (
	"io"
	"net/http"
	"strings"

	"github.com/square/quotaservice/config"
	pb "github.com/square/quotaservice/protos/config"
)

type namespacesAPIHandler struct {
	a Administrable
}

func newNamespacesAPIHandler(admin Administrable) (a *namespacesAPIHandler) {
	return &namespacesAPIHandler{a: admin}
}

func (a *namespacesAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ns := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api"), "/")
	user := getUsername(r)

	switch r.Method {
	case "GET":
		err := writeNamespace(a, w, ns)

		if err != nil {
			writeJSONError(w, err)
		}
	case "DELETE":
		if ns == "" {
			writeJSONError(w, &httpError{"", http.StatusNotFound})
			return
		}

		err := a.a.DeleteNamespace(ns, NewContext(user, TODO))

		if err != nil {
			writeJSONError(w, &httpError{err.Error(), http.StatusBadRequest})
		} else {
			writeJSONOk(w)
		}
	case "PUT":
		if ns == "" {
			writeJSONError(w, &httpError{"", http.StatusNotFound})
			return
		}

		changeNamespace(w, r, ns, func(c *pb.NamespaceConfig) error {
			return a.a.UpdateNamespace(c, NewContext(user, TODO))
		})
	case "POST":
		if ns == "" {
			updateConfig(a, w, r)
		} else {
			changeNamespace(w, r, ns, func(c *pb.NamespaceConfig) error {
				return a.a.AddNamespace(c, NewContext(user, TODO))
			})
		}
	default:
		writeJSONError(w, &httpError{"Unknown method " + r.Method, http.StatusBadRequest})
	}
}

func writeNamespace(a *namespacesAPIHandler, w http.ResponseWriter, namespace string) *httpError {
	var object interface{}
	cfgs := a.a.Configs()

	if namespace == "" || namespace == config.GlobalNamespace {
		object = cfgs
	} else {
		if _, exists := cfgs.Namespaces[namespace]; !exists {
			return &httpError{"Unable to locate namespace " + namespace, http.StatusNotFound}
		}

		object = cfgs.Namespaces[namespace]
	}

	writeJSON(w, object)
	return nil
}

func updateConfig(a *namespacesAPIHandler, w http.ResponseWriter, r *http.Request) {
	c := &pb.ServiceConfig{}
	e := unmarshalJSON(r.Body, c)

	if e != nil {
		writeJSONError(w, &httpError{e.Error(), http.StatusInternalServerError})
		return
	}

	e = a.a.UpdateConfig(c, NewContext(getUsername(r), TODO))

	if e != nil {
		writeJSONError(w, &httpError{e.Error(), http.StatusInternalServerError})
	} else {
		writeJSONOk(w)
	}
}

func changeNamespace(w http.ResponseWriter, r *http.Request, namespace string, updater func(*pb.NamespaceConfig) error) {
	c, e := getNamespaceConfig(r.Body)

	if e != nil {
		writeJSONError(w, &httpError{e.Error(), http.StatusInternalServerError})
		return
	}

	if c.Name == "" {
		c.Name = namespace
	}

	e = updater(c)

	if e != nil {
		writeJSONError(w, &httpError{e.Error(), http.StatusInternalServerError})
	} else {
		writeJSONOk(w)
	}
}

func getNamespaceConfig(r io.Reader) (*pb.NamespaceConfig, error) {
	c := &pb.NamespaceConfig{}
	err := unmarshalJSON(r, c)
	return c, err
}
