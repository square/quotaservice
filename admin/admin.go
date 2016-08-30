// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

// Package implements admin UIs and a REST API
package admin

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/maniksurtani/quotaservice/logging"
)

const (
	LogPattern = "%s - - [%s] \"%s\" %d %d %f\n"
)

type HttpError struct {
	message string
	status  int
}

type ResponseWrapper struct {
	http.ResponseWriter

	ip            string
	time          time.Time
	method        string
	uri           string
	protocol      string
	status        int
	responseBytes int64
	elapsedTime   time.Duration
}

// ServeAdminConsole serves up an admin console for an Administrable using Go's built-in HTTP server
// library. `assetsDirectory` contains HTML templates and other UI assets. If empty, no UI will be
// served, and only REST endpoints under `/api/` will be served.
func ServeAdminConsole(a Administrable, mux *http.ServeMux, assetsDirectory string, development bool) {
	if assetsDirectory != "" {
		msg := "Serving assets from %s"

		if development {
			msg += " (in development mode)"
		}

		logging.Printf(msg, assetsDirectory)

		mux.Handle("/", loggingHandler(http.RedirectHandler("/admin/", 301)))
		mux.Handle("/admin/", loggingHandler(NewUIHandler(a, assetsDirectory, development)))
		mux.Handle("/js/", loggingHandler(http.FileServer(http.Dir(assetsDirectory))))
		mux.Handle("/favicon.ico", http.NotFoundHandler())
	} else {
		logging.Print("Not serving admin web UI.")
		mux.Handle("/", loggingHandler(http.NotFoundHandler()))
	}

	bucketsHandler := NewBucketsAPIHandler(a)
	namespacesHandler := NewNamespacesAPIHandler(a)

	apiHandler := loggingHandler(jsonResponseHandler(apiRequestHandler(
		namespacesHandler, bucketsHandler)))

	mux.Handle("/api", apiHandler)
	mux.Handle("/api/", apiHandler)

	statsHandler := loggingHandler(jsonResponseHandler(NewStatsAPIHandler(a)))
	mux.Handle("/api/stats", statsHandler)
	mux.Handle("/api/stats/", statsHandler)
}

func (r *ResponseWrapper) Write(p []byte) (int, error) {
	bytes, err := r.ResponseWriter.Write(p)
	r.responseBytes += int64(bytes)
	return bytes, err
}

func (r *ResponseWrapper) Header() http.Header {
	return r.ResponseWriter.Header()
}

func (r *ResponseWrapper) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *ResponseWrapper) log() {
	timeFormatted := r.time.Format("02/Jan/2006 03:04:05")
	requestLine := fmt.Sprintf("%s %s %s", r.method, r.uri, r.protocol)
	logging.Printf(LogPattern,
		r.ip, timeFormatted, requestLine, r.status, r.responseBytes,
		r.elapsedTime.Seconds())
}

func jsonResponseHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func apiRequestHandler(namespacesHandler, bucketsHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := strings.SplitN(strings.Trim(r.URL.Path, "/"), "/", 3)

		// [api, {namespace}, {bucket}]
		if len(params) == 3 {
			bucketsHandler.ServeHTTP(w, r)
		} else {
			namespacesHandler.ServeHTTP(w, r)
		}
	})
}

func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr
		if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
			clientIP = clientIP[:colon]
		}

		response := &ResponseWrapper{
			ResponseWriter: w,
			ip:             clientIP,
			time:           time.Time{},
			method:         r.Method,
			uri:            r.RequestURI,
			protocol:       r.Proto,
			status:         http.StatusOK,
			elapsedTime:    time.Duration(0)}

		startTime := time.Now()
		next.ServeHTTP(response, r)
		finishTime := time.Now()

		response.time = finishTime.UTC()
		response.elapsedTime = finishTime.Sub(startTime)
		response.log()
	})
}
