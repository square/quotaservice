package admin

import (
	"bytes"
	"encoding/json"
	"github.com/maniksurtani/quotaservice/config"
	"reflect"
	"testing"
)

func TestExtractNamespace(t *testing.T) {
	ns, n := extractNamespaceName("ns/n")
	if ns != "ns" {
		t.Fatal("Expecting namespace 'ns'")
	}
	if n != "n" {
		t.Fatal("Expecting name 'n'")
	}
}

func TestUnmarshalBucketConfig(t *testing.T) {
	c := config.NewDefaultBucketConfig("Blah 123")
	c.FillRate = 12345
	c.MaxDebtMillis = 54321
	c.MaxIdleMillis = 67890
	c.MaxTokensPerRequest = 9876
	c.Size = 50000

	b, e := json.Marshal(c)
	if e != nil {
		t.Fatal("Unable to JSONify proto", e)
	}

	reRead, err := getBucketConfig(bytes.NewReader(b))
	if err != nil {
		t.Fatal("Unable to unmarshal JSON", err)
	}
	if !reflect.DeepEqual(c, reRead) {
		t.Fatalf("Two representations aren't equal: %+v != %+v", c, reRead)
	}
}

func TestUnmarshalNamespaceConfig(t *testing.T) {
	n := config.NewDefaultNamespaceConfig("Blah Namespace 123")
	n.MaxDynamicBuckets = 8000
	config.SetDynamicBucketTemplate(n, config.NewDefaultBucketConfig(""))

	c1 := config.NewDefaultBucketConfig("Blah 123")
	c1.FillRate = 12345
	c1.MaxDebtMillis = 54321
	c1.MaxIdleMillis = 67890
	c1.MaxTokensPerRequest = 9876
	c1.Size = 50000

	c2 := config.NewDefaultBucketConfig("Blah 456")
	c2.FillRate = 123450
	c2.MaxDebtMillis = 543210
	c2.MaxIdleMillis = 678900
	c2.MaxTokensPerRequest = 98760
	c2.Size = 5000

	c3 := config.NewDefaultBucketConfig("Blah 789")
	c3.FillRate = 1234500
	c3.MaxDebtMillis = 5432100
	c3.MaxIdleMillis = 6789000
	c3.MaxTokensPerRequest = 987600
	c3.Size = 500

	config.AddBucket(n, c1)
	config.AddBucket(n, c2)
	config.AddBucket(n, c3)

	b, e := json.Marshal(n)
	if e != nil {
		t.Fatal("Unable to JSONify proto", e)
	}

	reRead, err := getNamespaceConfig(bytes.NewReader(b))
	if err != nil {
		t.Fatal("Unable to unmarshal JSON", err)
	}

	if !reflect.DeepEqual(n, reRead) {
		t.Fatalf("Two representations aren't equal: %+v != %+v", n, reRead)
	}
}
