// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package admin

import (
	"net/http"
	"strings"

	"github.com/maniksurtani/quotaservice/stats"
)

type statsAPIHandler struct {
	a Administrable
}

type bucketStats struct {
	Ns     string               `json:"namespace"`
	Hits   []*stats.BucketScore `json:"topHits"`
	Misses []*stats.BucketScore `json:"topMisses"`
}

func NewStatsAPIHandler(admin Administrable) (a *statsAPIHandler) {
	return &statsAPIHandler{a: admin}
}

func (a *statsAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ns := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/stats"), "/")

	if r.Method != "GET" {
		writeJSONError(w, &HttpError{"Unknown method " + r.Method, http.StatusBadRequest})
		return
	}

	if ns == "" {
		writeJSONError(w, &HttpError{"No namespace specified", http.StatusBadRequest})
		return
	}

	err := writeStats(a, w, ns)

	if err != nil {
		writeJSONError(w, err)
	}
}

func writeStats(a *statsAPIHandler, w http.ResponseWriter, path string) *HttpError {
	params := strings.SplitN(path, "/", 2)
	namespace := params[0]

	cfgs := a.a.Configs()

	if _, exists := cfgs.Namespaces[namespace]; !exists {
		return &HttpError{"Unable to locate namespace " + namespace, http.StatusNotFound}
	}

	if len(params) == 2 {
		stat := a.a.DynamicBucketStats(namespace, params[1])

		if stat == nil {
			return &HttpError{"No stats listener configured", http.StatusBadRequest}
		}

		responseMap := make(map[string]*stats.BucketScores)
		responseMap[params[1]] = stat
		writeJSON(w, responseMap)
		return nil
	}

	hits := a.a.TopDynamicHits(namespace)

	if hits == nil {
		return &HttpError{"No stats listener configured", http.StatusBadRequest}
	}

	misses := a.a.TopDynamicMisses(namespace)

	if misses == nil {
		return &HttpError{"No stats listener configured", http.StatusBadRequest}
	}

	writeJSON(w, &bucketStats{namespace, hits, misses})

	return nil
}
