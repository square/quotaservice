// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

// Package config implements configs for the quotaservice
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
	"github.com/square/quotaservice/logging"
	pb "github.com/square/quotaservice/protos/config"
	"gopkg.in/yaml.v2"
)

const (
	GlobalNamespace           = "___GLOBAL___"
	DefaultBucketName         = "___DEFAULT_BUCKET___"
	DynamicBucketTemplateName = "___DYNAMIC_BUCKET_TPL___"
	initialVersion            = 0
	initialHash               = "___INITIAL_HASH___"
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
		Version:             initialVersion}
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
	persister := NewMemoryConfigPersister()
	if err := persister.PersistAndNotify(initialHash, p); err != nil {
		logging.Fatalf("Unable to persist initial configuration: %v", err)
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

func UnmarshalBytes(b []byte) (*pb.ServiceConfig, error) {
	p := &pb.ServiceConfig{}
	e := proto.Unmarshal(b, p)
	return p, e
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

func DifferentBucketConfigs(c1, c2 *pb.BucketConfig) bool {
	if c1 == nil && c2 == nil {
		// Both are nil - so not different
		return false
	}

	if c1 == nil || c2 == nil {
		// One of them is NOT nil!
		return true
	}

	return c1.Name != c2.Name ||
		c1.Namespace != c2.Namespace ||
		c1.Size != c2.Size ||
		c1.FillRate != c2.FillRate ||
		c1.WaitTimeoutMillis != c2.WaitTimeoutMillis ||
		c1.MaxIdleMillis != c2.MaxIdleMillis ||
		c1.MaxDebtMillis != c2.MaxDebtMillis ||
		c1.MaxTokensPerRequest != c2.MaxTokensPerRequest
}

func DifferentNamespaceConfigs(c1, c2 *pb.NamespaceConfig) bool {
	different := c1.Name != c2.Name ||
		c1.MaxDynamicBuckets != c2.MaxDynamicBuckets ||
		DifferentBucketConfigs(c1.DefaultBucket, c2.DefaultBucket) ||
		DifferentBucketConfigs(c1.DynamicBucketTemplate, c2.DynamicBucketTemplate) ||
		len(c1.Buckets) != len(c2.Buckets)

	if different {
		return true
	}

	// Now check named buckets
	for name, b1 := range c1.Buckets {
		b2, exists := c2.Buckets[name]
		if !exists || DifferentBucketConfigs(b1, b2) {
			// Short-circuit
			return true
		}
	}

	return false
}

func cloneConfig(cfg *pb.ServiceConfig) *pb.ServiceConfig {
	return proto.Clone(cfg).(*pb.ServiceConfig)
}

func cloneConfigs(cfgs map[string]*pb.ServiceConfig) []*pb.ServiceConfig {
	cloned := make([]*pb.ServiceConfig, 0, len(cfgs))

	for _, v := range cfgs {
		cloned = append(cloned, cloneConfig(v))
	}

	return cloned
}
