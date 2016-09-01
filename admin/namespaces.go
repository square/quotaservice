// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package admin

import (
	"io"
	"net/http"
	"strings"

	"github.com/maniksurtani/quotaservice/config"
	pb "github.com/maniksurtani/quotaservice/protos/config"
)

type namespacesAPIHandler struct {
	a Administrable
}

func NewNamespacesAPIHandler(admin Administrable) (a *namespacesAPIHandler) {
	return &namespacesAPIHandler{a: admin}
}

func (a *namespacesAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ns := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api"), "/")

	switch r.Method {
	case "GET":
		err := writeNamespace(a, w, ns)

		if err != nil {
			writeJSONError(w, err)
		}
	case "DELETE":
		err := a.a.DeleteNamespace(ns)

		if err != nil {
			writeJSONError(w, &HttpError{err.Error(), http.StatusBadRequest})
		} else {
			writeJSONOk(w)
		}
	case "PUT":
		changeNamespace(w, r, func(c *pb.NamespaceConfig) error {
			return a.a.UpdateNamespace(c)
		})
	case "POST":
		if ns == "" {
			updateConfig(a, w, r)
		} else {
			changeNamespace(w, r, func(c *pb.NamespaceConfig) error {
				return a.a.AddNamespace(c)
			})
		}
	default:
		writeJSONError(w, &HttpError{"Unknown method " + r.Method, http.StatusBadRequest})
	}
}

func writeNamespace(a *namespacesAPIHandler, w http.ResponseWriter, namespace string) *HttpError {
	var object interface{}
	cfgs := a.a.Configs()

	if namespace == "" || namespace == config.GlobalNamespace {
		object = cfgs
	} else {
		if _, exists := cfgs.Namespaces[namespace]; !exists {
			return &HttpError{"Unable to locate namespace " + namespace, http.StatusNotFound}
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
		writeJSONError(w, &HttpError{e.Error(), http.StatusInternalServerError})
		return
	}

	e = a.a.UpdateConfig(c, getUsername(r))

	if e != nil {
		writeJSONError(w, &HttpError{e.Error(), http.StatusInternalServerError})
	} else {
		err := writeNamespace(a, w, "")

		if err != nil {
			writeJSONError(w, err)
		}
	}
}

func changeNamespace(w http.ResponseWriter, r *http.Request, updater func(*pb.NamespaceConfig) error) {
	c, e := getNamespaceConfig(r.Body)

	if e != nil {
		writeJSONError(w, &HttpError{e.Error(), http.StatusInternalServerError})
		return
	}

	e = updater(c)

	if e != nil {
		writeJSONError(w, &HttpError{e.Error(), http.StatusInternalServerError})
	} else {
		writeJSONOk(w)
	}
}

func getNamespaceConfig(r io.Reader) (*pb.NamespaceConfig, error) {
	c := &pb.NamespaceConfig{}
	err := unmarshalJSON(r, c)
	return c, err
}
