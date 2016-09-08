// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package admin

import (
	"net/http"

	pb "github.com/maniksurtani/quotaservice/protos/config"
)

type configsAPIHandler struct {
	a Administrable
}

func NewConfigsAPIHandler(admin Administrable) (a *configsAPIHandler) {
	return &configsAPIHandler{a: admin}
}

type configsResponse struct {
	Configs []*pb.ServiceConfig `json:"configs"`
}

func (a *configsAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeJSONError(w, &HttpError{"Unknown method " + r.Method, http.StatusBadRequest})
		return
	}

	configs, err := a.a.HistoricalConfigs()

	if err != nil {
		writeJSONError(w, &HttpError{"Error reading configs " + err.Error(), http.StatusInternalServerError})
	} else {
		writeJSON(w, &configsResponse{configs})
	}
}
