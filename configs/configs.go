// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

// Package implements configs for the quotaservice
package configs

import (
	"fmt"
	"io"
	"io/ioutil"
	"github.com/maniksurtani/quotaservice/logging"
	"gopkg.in/yaml.v2"
)

type ServiceConfig struct {
	MetricsEnabled      bool                        `yaml:"metrics_enabled"`
	GlobalDefaultBucket *BucketConfig               `yaml:"global_default_bucket,flow"`
	Namespaces          map[string]*NamespaceConfig `yaml:",flow"`
}

type NamespaceConfig struct {
	DefaultBucket         *BucketConfig            `yaml:"default_bucket,flow"`
	DynamicBucketTemplate *BucketConfig            `yaml:"dynamic_bucket_template,flow"`
	MaxDynamicBuckets     int                      `yaml:"max_dynamic_buckets"`
	Buckets               map[string]*BucketConfig `yaml:",flow"`
}

type BucketConfig struct {
	Size                int64
	FillRate            int64 `yaml:"fill_rate"`
	WaitTimeoutMillis   int64 `yaml:"wait_timeout_millis"`
	MaxIdleMillis       int64 `yaml:"max_idle_millis"`
	MaxDebtMillis       int64 `yaml:"max_debt_millis"`
	MaxTokensPerRequest int64 `yaml:"max_tokens_per_request"`
}

func (b *BucketConfig) String() string {
	return fmt.Sprint(*b)
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

	return ApplyDefaults(cfg)
}

func ApplyDefaults(cfg *ServiceConfig) *ServiceConfig {
	applyBucketDefaults(cfg.GlobalDefaultBucket)

	for name, ns := range cfg.Namespaces {
		if ns.DefaultBucket != nil && ns.DynamicBucketTemplate != nil {
			panic(fmt.Sprintf("Namespace %v is not allowed to have a default bucket as well as allow dynamic buckets.", name))
		}

		// Ensure the namespace's bucket map exists.
		if ns.Buckets == nil {
			ns.Buckets = make(map[string]*BucketConfig)
		}

		applyBucketDefaults(ns.DefaultBucket)
		applyBucketDefaults(ns.DynamicBucketTemplate)

		for _, b := range ns.Buckets {
			applyBucketDefaults(b)
		}
	}

	logging.Printf("Read config %+v", cfg)
	return cfg
}

func NewDefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		MetricsEnabled:        true,
		GlobalDefaultBucket:   NewDefaultBucketConfig(),
		Namespaces:            make(map[string]*NamespaceConfig)}
}

func NewDefaultNamespaceConfig() *NamespaceConfig {
	return &NamespaceConfig{Buckets: make(map[string]*BucketConfig)}
}

func NewDefaultBucketConfig() *BucketConfig {
	return &BucketConfig{Size: 100, FillRate: 50, WaitTimeoutMillis: 1000, MaxIdleMillis: -1, MaxDebtMillis: 10000}
}

func applyBucketDefaults(b *BucketConfig) {
	if b != nil {
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
}
