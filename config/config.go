// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

// Package implements configs for the quotaservice
package config

import (
	"io"
	"io/ioutil"

	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	pb "github.com/maniksurtani/quotaservice/protos/config"
	"gopkg.in/yaml.v2"
)

const (
	GlobalNamespace           = "___GLOBAL___"
	DefaultBucketName         = "___DEFAULT_BUCKET___"
	DynamicBucketTemplateName = "___DYNAMIC_BUCKET_TPL___"
	initial_version           = 0
)

func ApplyDefaults(sc *pb.ServiceConfig) {
	if sc.GlobalDefaultBucket != nil {
		ApplyBucketDefaults(sc.GlobalDefaultBucket)
		sc.GlobalDefaultBucket.Name = DefaultBucketName
	}

	for name, ns := range sc.Namespaces {
		ns.Name = name
		if ns.DefaultBucket != nil && ns.DynamicBucketTemplate != nil {
			panic(fmt.Sprintf("Namespace %v is not allowed to have a default bucket as well as allow dynamic buckets.", name))
		}

		// Ensure the namespace's bucket map exists.
		if ns.Buckets == nil {
			ns.Buckets = make(map[string]*pb.BucketConfig)
		}

		if ns.DefaultBucket != nil {
			ApplyBucketDefaults(ns.DefaultBucket)
			ns.DefaultBucket.Name = DefaultBucketName
			ns.DefaultBucket.Namespace = ns.Name
		}

		if ns.DynamicBucketTemplate != nil {
			ApplyBucketDefaults(ns.DynamicBucketTemplate)
			ns.DynamicBucketTemplate.Name = DynamicBucketTemplateName
			ns.DynamicBucketTemplate.Namespace = ns.Name
		}

		for n, b := range ns.Buckets {
			ApplyBucketDefaults(b)
			b.Name = n
			b.Namespace = ns.Name
		}
	}
}

func NamespaceNames(sc *pb.ServiceConfig) []string {
	if sc.Namespaces == nil || len(sc.Namespaces) == 0 {
		return []string{}
	}

	names := make([]string, 0, len(sc.Namespaces))
	for ns := range sc.Namespaces {
		names = append(names, ns)
	}

	return names
}

func ApplyBucketDefaults(b *pb.BucketConfig) {
	if b.Size == 0 {
		b.Size = 100
	}

	if b.FillRate == 0 {
		b.FillRate = 50
	}

	if b.WaitTimeoutMillis == 0 {
		b.WaitTimeoutMillis = 1000
	}

	if b.MaxIdleMillis == 0 {
		b.MaxIdleMillis = -1
	}

	if b.MaxDebtMillis == 0 {
		b.MaxDebtMillis = 10000
	}

	if b.MaxTokensPerRequest == 0 {
		b.MaxTokensPerRequest = b.FillRate
	}
}

func FQN(b *pb.BucketConfig) string {
	if b.Namespace == "" {
		// This is a global default.
		return FullyQualifiedName(GlobalNamespace, DefaultBucketName)
	}

	return FullyQualifiedName(b.Namespace, b.Name)
}

func ReadConfigFromFile(filename string) *pb.ServiceConfig {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("Unable to open file %v. Error: %v", filename, err))
	}

	return readConfigFromBytes(bytes)
}

func ReadConfig(yamlStream io.Reader) *pb.ServiceConfig {
	bytes, err := ioutil.ReadAll(yamlStream)
	if err != nil {
		panic(fmt.Sprintf("Unable to open reader. Error: %v", err))
	}

	return readConfigFromBytes(bytes)
}

func readConfigFromBytes(bytes []byte) *pb.ServiceConfig {
	cfg := NewDefaultServiceConfig()
	cfg.GlobalDefaultBucket = nil
	if err := yaml.Unmarshal(bytes, cfg); err != nil {
		panic(fmt.Sprintf("Unable to read YAML. Error: %v", err))
	}

	ApplyDefaults(cfg)
	return cfg
}

func NewDefaultServiceConfig() *pb.ServiceConfig {
	return &pb.ServiceConfig{
		GlobalDefaultBucket: nil,
		Namespaces:          make(map[string]*pb.NamespaceConfig),
		User:                "quotaservice",
		Date:                time.Now().Unix(),
		Version:             initial_version}
}

func NewDefaultNamespaceConfig(name string) *pb.NamespaceConfig {
	return &pb.NamespaceConfig{
		Buckets: make(map[string]*pb.BucketConfig),
		Name:    name}
}

func NewDefaultBucketConfig(name string) *pb.BucketConfig {
	return &pb.BucketConfig{
		Size:              100,
		FillRate:          50,
		WaitTimeoutMillis: 1000,
		MaxIdleMillis:     -1,
		MaxDebtMillis:     10000,
		Name:              name}
}

func FromJSON(j []byte) (*pb.ServiceConfig, error) {
	p := &pb.ServiceConfig{}
	e := json.Unmarshal(j, p)
	if e != nil {
		return nil, e
	}

	return p, nil
}

func NamespaceFromJSON(j []byte) (*pb.NamespaceConfig, error) {
	p := &pb.NamespaceConfig{}
	e := json.Unmarshal(j, p)
	if e != nil {
		return nil, e
	}

	return p, nil
}

func FullyQualifiedName(namespace, bucketName string) string {
	return namespace + ":" + bucketName
}

func NewMemoryConfig(p *pb.ServiceConfig) ConfigPersister {
	marshalled, e := Marshal(p)
	if e != nil {
		panic(e)
	}

	persister := NewMemoryConfigPersister()
	if err := persister.PersistAndNotify(marshalled); err != nil {
		panic(err)
	}

	return persister
}

func Marshal(p *pb.ServiceConfig) (io.Reader, error) {
	b, e := proto.Marshal(p)
	if e != nil {
		return nil, e
	}

	return bytes.NewReader(b), nil
}

func Unmarshal(r io.Reader) (*pb.ServiceConfig, error) {
	b, e := ioutil.ReadAll(r)
	if e != nil {
		return nil, e
	}

	p := &pb.ServiceConfig{}
	e = proto.Unmarshal(b, p)
	if e != nil {
		return nil, e
	}

	return p, nil
}

func AddBucket(n *pb.NamespaceConfig, b *pb.BucketConfig) error {
	if b.Name == "" {
		return errors.New("Bucket name cannot be nil or empty.")
	}
	n.Buckets[b.Name] = b
	b.Namespace = n.Name
	return nil
}

func SetDynamicBucketTemplate(n *pb.NamespaceConfig, b *pb.BucketConfig) {
	b.Name = DynamicBucketTemplateName
	b.Namespace = n.Name
	n.DynamicBucketTemplate = b
}

func AddNamespace(s *pb.ServiceConfig, n *pb.NamespaceConfig) error {
	if n.Name == "" {
		return errors.New("Namespace name cannot be nil or empty.")
	}
	s.Namespaces[n.Name] = n
	return nil
}
