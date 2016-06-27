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

type bucketsAPIHandler struct {
	a Administrable
}

func NewBucketsAPIHandler(admin Administrable) (a *bucketsAPIHandler) {
	return &bucketsAPIHandler{a: admin}
}

func (a *bucketsAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := strings.SplitN(strings.Trim(r.URL.Path, "/"), "/", 3)
	namespace, bucket := params[1], params[2]

	switch r.Method {
	case "GET":
		err := writeBucket(a, w, namespace, bucket)

		if err != nil {
			writeJSONError(w, err)
		}
	case "DELETE":
		err := a.a.DeleteBucket(namespace, bucket)

		if err != nil {
			writeJSONError(w, &HttpError{err.Error(), http.StatusBadRequest})
		}
	case "PUT":
		changeBucket(w, r, func(c *pb.BucketConfig) error {
			return a.a.UpdateBucket(namespace, c)
		})
	case "POST":
		changeBucket(w, r, func(c *pb.BucketConfig) error {
			return a.a.AddBucket(namespace, c)
		})
	default:
		writeJSONError(w, &HttpError{"Unknown method " + r.Method, http.StatusBadRequest})
	}
}

func changeBucket(w http.ResponseWriter, r *http.Request, updater func(*pb.BucketConfig) error) {
	c, e := getBucketConfig(r.Body)

	if e != nil {
		writeJSONError(w, &HttpError{e.Error(), http.StatusInternalServerError})
		return
	}

	e = updater(c)

	if e != nil {
		writeJSONError(w, &HttpError{e.Error(), http.StatusInternalServerError})
	}
}

func getBucketConfig(r io.Reader) (*pb.BucketConfig, error) {
	c := &pb.BucketConfig{}
	config.ApplyBucketDefaults(c)
	err := unmarshalJSON(r, c)
	return c, err
}

func writeBucket(a *bucketsAPIHandler, w http.ResponseWriter, namespace, bucket string) *HttpError {
	namespaceConfig := a.a.Configs().Namespaces[namespace]

	if namespaceConfig == nil {
		return &HttpError{"Unable to locate namespace " + namespace, http.StatusNotFound}
	}

	if bucket == "" {
		// this shouldn't really be possible
		return &HttpError{"No bucket given", http.StatusNotFound}
	}

	bucketConfig := namespaceConfig.Buckets[bucket]

	if bucketConfig == nil {
		return &HttpError{"Unable to locate bucket " + bucket + " in namespace " + namespace, http.StatusNotFound}
	}

	writeJSON(w, bucketConfig)
	return nil
}
