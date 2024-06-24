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

type bucketsAPIHandler struct {
	a Administrable
}

func newBucketsAPIHandler(admin Administrable) (a *bucketsAPIHandler) {
	return &bucketsAPIHandler{a: admin}
}

func (a *bucketsAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := strings.SplitN(strings.Trim(r.URL.Path, "/"), "/", 3)
	namespace, bucket := params[1], params[2]
	user := getUsername(r)

	switch r.Method {
	case "GET":
		err := writeBucket(a, w, namespace, bucket)

		if err != nil {
			writeJSONError(w, err)
		}
	case "DELETE":
		err := a.a.DeleteBucket(namespace, bucket, user)

		if err != nil {
			writeJSONError(w, &httpError{err.Error(), http.StatusBadRequest})
		} else {
			writeJSONOk(w)
		}
	case "PUT":
		changeBucket(w, r, bucket, func(c *pb.BucketConfig) error {
			return a.a.UpdateBucket(namespace, c, user)
		})
	case "POST":
		changeBucket(w, r, bucket, func(c *pb.BucketConfig) error {
			return a.a.AddBucket(namespace, c, user)
		})
	default:
		writeJSONError(w, &httpError{"Unknown method " + r.Method, http.StatusBadRequest})
	}
}

func changeBucket(w http.ResponseWriter, r *http.Request, bucket string, updater func(*pb.BucketConfig) error) {
	c, e := getBucketConfig(r.Body)

	if e != nil {
		writeJSONError(w, &httpError{e.Error(), http.StatusInternalServerError})
		return
	}

	if c.Name == "" {
		c.Name = bucket
	}

	e = updater(c)

	if e != nil {
		writeJSONError(w, &httpError{e.Error(), http.StatusInternalServerError})
	} else {
		writeJSONOk(w)
	}
}

func getBucketConfig(r io.Reader) (*pb.BucketConfig, error) {
	c := &pb.BucketConfig{}
	config.ApplyBucketDefaults(c)
	err := unmarshalJSON(r, c)
	return c, err
}

func writeBucket(a *bucketsAPIHandler, w http.ResponseWriter, namespace, bucket string) *httpError {
	config := a.a.Configs()
	namespaceConfig, exists := config.Namespaces[namespace]

	if !exists {
		return &httpError{"Unable to locate namespace " + namespace, http.StatusNotFound}
	}

	if bucket == "" {
		// this shouldn't really be possible
		return &httpError{"No bucket given", http.StatusNotFound}
	}

	bucketConfig, exists := namespaceConfig.Buckets[bucket]

	if !exists {
		return &httpError{"Unable to locate bucket " + bucket + " in namespace " + namespace, http.StatusNotFound}
	}

	writeJSON(w, bucketConfig)
	return nil
}
