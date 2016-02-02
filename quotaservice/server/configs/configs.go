package configs
import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"fmt"
	"github.com/maniksurtani/qs/quotaservice/server/logging"
)

type Configs struct {
	Port                  int
	AdminPort             int `yaml:"admin_port"`
	MetricsEnabled        bool `yaml:"metrics_enabled"`
	FillerFrequencyMillis int `yaml:"filler_frequency_millis"`
	UseDefaultBuckets     bool `yaml:"use_default_buckets"`
	Buckets               map[string]*BucketConfig `yaml:",flow"`
}

type BucketConfig struct {
	Size              int
	FillRate          int `yaml:"fill_rate"`
	WaitTimeoutMillis int `yaml:"wait_timeout_millis"`
	RejectIfEmpty     bool `yaml:"reject_if_empty"`
}

func (b *BucketConfig) String() string {
	return fmt.Sprint(*b)
}

func ReadConfig(filename string) *Configs {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		logging.Fatalf("Unable to open config file %v. Error: %v", filename, err)
		panic(fmt.Sprintf("Unable to open config file %v. Error: %v", filename, err))
	}

	logging.Print(string(dat))
	cfg := NewDefaultConfig()
	yaml.Unmarshal(dat, cfg)

	// Apply defaults to all named buckets
	// TODO(manik) there must be a better way to apply defaults when parsing YAML!
	for _, b := range cfg.Buckets {
		applyBucketDefaults(b)
	}

	logging.Printf("Read config %+v", cfg)
	return cfg
}

func NewDefaultConfig() *Configs {
	return &Configs{
		Port: 11100,
		AdminPort: 8080,
		MetricsEnabled: false,
		FillerFrequencyMillis: 1000,
		UseDefaultBuckets: false,
	}
}

func applyBucketDefaults(b *BucketConfig) {
	if b.FillRate == 0 {
		b.FillRate = 100
	}

	if b.Size == 0 {
		b.Size = 1000
	}

	if b.WaitTimeoutMillis == 0 {
		b.WaitTimeoutMillis = 5000
	}
}
