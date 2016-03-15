// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

// Package implements configs for the quotaservice
package quotaservice

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/maniksurtani/quotaservice/logging"
	"gopkg.in/yaml.v2"
	pb "github.com/maniksurtani/quotaservice/protos/config"
	"github.com/golang/protobuf/proto"
)

type ServiceConfig struct {
	GlobalDefaultBucket *BucketConfig               `yaml:"global_default_bucket,flow"`
	Namespaces          map[string]*NamespaceConfig `yaml:",flow"`
	Version             int
}

func (s *ServiceConfig) String() string {
	return fmt.Sprintf("ServiceConfig{default: %v, namespaces: %v}",
		s.GlobalDefaultBucket, s.Namespaces)
}

func (s *ServiceConfig) AddNamespace(namespace string, n *NamespaceConfig) *ServiceConfig {
	s.Namespaces[namespace] = n
	n.Name = namespace
	return s
}

func (s *ServiceConfig) ToProto() *pb.ServiceConfig {
	return &pb.ServiceConfig{
		Version: proto.Int(s.Version),
		GlobalDefaultBucket: bucketToProto(defaultBucketName, s.GlobalDefaultBucket),
		Namespaces: namespaceMapToProto(s.Namespaces)}
}

func (s *ServiceConfig) ApplyDefaults() *ServiceConfig {
	if s.GlobalDefaultBucket != nil {
		s.GlobalDefaultBucket.ApplyDefaults()
		s.GlobalDefaultBucket.Name = defaultBucketName
	}

	for name, ns := range s.Namespaces {
		ns.Name = name
		if ns.DefaultBucket != nil && ns.DynamicBucketTemplate != nil {
			panic(fmt.Sprintf("Namespace %v is not allowed to have a default bucket as well as allow dynamic buckets.", name))
		}

		// Ensure the namespace's bucket map exists.
		if ns.Buckets == nil {
			ns.Buckets = make(map[string]*BucketConfig)
		}

		if ns.DefaultBucket != nil {
			ns.DefaultBucket.ApplyDefaults()
			ns.DefaultBucket.Name = defaultBucketName
			ns.DefaultBucket.namespace = ns
		}

		if ns.DynamicBucketTemplate != nil {
			ns.DynamicBucketTemplate.ApplyDefaults()
			ns.DynamicBucketTemplate.Name = dynamicBucketTemplateName
			ns.DynamicBucketTemplate.namespace = ns
		}

		for n, b := range ns.Buckets {
			b.ApplyDefaults()
			b.Name = n
			b.namespace = ns
		}
	}

	return s
}

func (s *ServiceConfig) NamespaceNames() []string {
	if s.Namespaces == nil || len(s.Namespaces) == 0 {
		return []string{}
	}

	names := make([]string, len(s.Namespaces))
	i := 0
	for ns, _ := range s.Namespaces {
		names[i] = ns
		i++
	}

	return names
}

type NamespaceConfig struct {
	DefaultBucket         *BucketConfig            `yaml:"default_bucket,flow"`
	DynamicBucketTemplate *BucketConfig            `yaml:"dynamic_bucket_template,flow"`
	MaxDynamicBuckets     int                      `yaml:"max_dynamic_buckets"`
	Buckets               map[string]*BucketConfig `yaml:",flow"`
	Name                  string
}

func (n *NamespaceConfig) AddBucket(name string, b *BucketConfig) *NamespaceConfig {
	n.Buckets[name] = b
	b.Name = name
	return n
}

func (n *NamespaceConfig) ToProto() *pb.NamespaceConfig {
	return &pb.NamespaceConfig{
		DefaultBucket: bucketToProto(defaultBucketName, n.DefaultBucket),
		DynamicBucketTemplate: bucketToProto(dynamicBucketTemplateName, n.DynamicBucketTemplate),
		MaxDynamicBuckets: proto.Int(n.MaxDynamicBuckets),
		Buckets: bucketMapToProto(n.Buckets),
		Name: proto.String(n.Name)}
}

type BucketConfig struct {
	Size                int64
	FillRate            int64 `yaml:"fill_rate"`
	WaitTimeoutMillis   int64 `yaml:"wait_timeout_millis"`
	MaxIdleMillis       int64 `yaml:"max_idle_millis"`
	MaxDebtMillis       int64 `yaml:"max_debt_millis"`
	MaxTokensPerRequest int64 `yaml:"max_tokens_per_request"`
	namespace           *NamespaceConfig
	Name                string
}

func (b *BucketConfig) String() string {
	return fmt.Sprint(*b)
}

func (b *BucketConfig) ToProto() *pb.BucketConfig {
	return &pb.BucketConfig{
		Size: proto.Int64(b.Size),
		FillRate: proto.Int64(b.FillRate),
		WaitTimeoutMillis: proto.Int64(b.WaitTimeoutMillis),
		MaxIdleMillis: proto.Int64(b.MaxIdleMillis),
		MaxDebtMillis: proto.Int64(b.MaxDebtMillis),
		MaxTokensPerRequest: proto.Int64(b.MaxTokensPerRequest),
		Name: proto.String(b.Name)}
}

func (b *BucketConfig) ApplyDefaults() *BucketConfig {
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

	return b
}

func (b *BucketConfig) FQN() string {
	if b.namespace == nil {
		// This is a global default.
		return FullyQualifiedName(globalNamespace, defaultBucketName)
	}

	return FullyQualifiedName(b.namespace.Name, b.Name)
}

func ReadConfigFromFile(filename string) *ServiceConfig {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("Unable to open file %v. Error: %v", filename, err))
	}

	return readConfigFromBytes(bytes)
}

func ReadConfig(yamlStream io.Reader) *ServiceConfig {
	bytes, err := ioutil.ReadAll(yamlStream)
	if err != nil {
		panic(fmt.Sprintf("Unable to open reader. Error: %v", err))
	}

	return readConfigFromBytes(bytes)
}

func readConfigFromBytes(bytes []byte) *ServiceConfig {
	logging.Print(string(bytes))
	cfg := NewDefaultServiceConfig()
	cfg.GlobalDefaultBucket = nil
	yaml.Unmarshal(bytes, cfg)

	return cfg.ApplyDefaults()
}

func NewDefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		NewDefaultBucketConfig(),
		make(map[string]*NamespaceConfig),
		0}
}

func NewDefaultNamespaceConfig() *NamespaceConfig {
	return &NamespaceConfig{Buckets: make(map[string]*BucketConfig)}
}

func NewDefaultBucketConfig() *BucketConfig {
	return &BucketConfig{Size: 100, FillRate: 50, WaitTimeoutMillis: 1000, MaxIdleMillis: -1, MaxDebtMillis: 10000}
}

// Helpers to read to and write from proto representations
func bucketToProto(name string, b *BucketConfig) *pb.BucketConfig {
	if b == nil {
		return nil
	}

	return b.ToProto()
}

func bucketMapToProto(buckets map[string]*BucketConfig) []*pb.BucketConfig {
	slice := make([]*pb.BucketConfig, len(buckets))
	i := 0
	for n, b := range buckets {
		slice[i] = bucketToProto(n, b)
		i++
	}

	return slice
}

func namespaceMapToProto(namespaces map[string]*NamespaceConfig) []*pb.NamespaceConfig {
	slice := make([]*pb.NamespaceConfig, len(namespaces))
	i := 0
	for _, nsp := range namespaces {
		slice[i] = nsp.ToProto()
		i++
	}

	return slice
}

func FromProto(cfg *pb.ServiceConfig) *ServiceConfig {
	globalBucket := bucketFromProto(cfg.GlobalDefaultBucket, nil)
	return &ServiceConfig{
		GlobalDefaultBucket: globalBucket,
		Version: int(cfg.GetVersion()),
		Namespaces: namespacesFromProto(cfg.Namespaces)}
}

func bucketsFromProto(cfgs []*pb.BucketConfig, nsc *NamespaceConfig) map[string]*BucketConfig {
	buckets := make(map[string]*BucketConfig, len(cfgs))
	for _, cfg := range cfgs {
		b := bucketFromProto(cfg, nsc)
		if b != nil {
			buckets[b.Name] = b
		}
	}

	return buckets
}

func bucketFromProto(cfg *pb.BucketConfig, nsc *NamespaceConfig) (b *BucketConfig) {
	if cfg == nil {
		return
	}

	b = &BucketConfig{
		Size: cfg.GetSize(),
		FillRate: cfg.GetFillRate(),
		WaitTimeoutMillis: cfg.GetWaitTimeoutMillis(),
		MaxIdleMillis: cfg.GetMaxIdleMillis(),
		MaxDebtMillis: cfg.GetMaxDebtMillis(),
		MaxTokensPerRequest: cfg.GetMaxTokensPerRequest(),
		namespace: nsc, Name: cfg.GetName()}
	return
}

func namespacesFromProto(cfgs []*pb.NamespaceConfig) map[string]*NamespaceConfig {
	namespaces := make(map[string]*NamespaceConfig, len(cfgs))

	for _, cfg := range cfgs {
		ns := namespaceFromProto(cfg)
		if ns != nil {
			namespaces[ns.Name] = ns
		}
	}

	return namespaces
}

func namespaceFromProto(cfg *pb.NamespaceConfig) (n *NamespaceConfig) {
	if cfg == nil {
		return
	}

	n = &NamespaceConfig{
		MaxDynamicBuckets: int(cfg.GetMaxDynamicBuckets()),
		Name: cfg.GetName()}

	n.DefaultBucket = bucketFromProto(cfg.DefaultBucket, n)
	n.DynamicBucketTemplate = bucketFromProto(cfg.DynamicBucketTemplate, n)
	n.Buckets = bucketsFromProto(cfg.Buckets, n)

	return
}
