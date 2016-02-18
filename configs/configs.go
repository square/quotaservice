/*
 *   Copyright 2016 Manik Surtani
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package configs
import (
	"io/ioutil"
	"fmt"
	"github.com/maniksurtani/quotaservice/logging"
	"gopkg.in/yaml.v2"
	"io"
)

type ServiceConfig struct {
	AdminPort             int `yaml:"admin_port"`
	MetricsEnabled        bool `yaml:"metrics_enabled"`
	FillerFrequencyMillis int `yaml:"filler_frequency_millis"`
	GlobalDefaultBucket   *BucketConfig `yaml:"global_default_bucket,flow"`
	Namespaces            map[string]*NamespaceConfig `yaml:",flow"`
}

type NamespaceConfig struct {
	DefaultBucket         *BucketConfig `yaml:"default_bucket,flow"`
	DynamicBucketTemplate *BucketConfig `yaml:"dynamic_bucket_template,flow"`
	MaxDynamicBuckets     int `yaml:"max_dynamic_buckets"`
	Buckets               map[string]*BucketConfig `yaml:",flow"`
}

type BucketConfig struct {
	Size              int
	FillRate          int `yaml:"fill_rate"`
	WaitTimeoutMillis int `yaml:"wait_timeout_millis"`
	MaxIdleMillis     int `yaml:"max_idle_millis"`
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

	// Apply defaults
	// TODO(manik) there must be a better way to apply defaults when parsing YAML!
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
		AdminPort: 8080,
		MetricsEnabled: true,
		FillerFrequencyMillis: 1000,
		GlobalDefaultBucket: NewDefaultBucketConfig(),
		Namespaces: make(map[string]*NamespaceConfig)}
}

func NewDefaultNamespaceConfig() *NamespaceConfig {
	return &NamespaceConfig{Buckets: make(map[string]*BucketConfig)}
}

func NewDefaultBucketConfig() *BucketConfig {
	return &BucketConfig{Size: 100, FillRate: 50, WaitTimeoutMillis: 1000, MaxIdleMillis: -1}
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
	}
}
