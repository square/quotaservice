// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package metrics

import (
	"github.com/codahale/hdrhistogram"
)

type metrics struct {
	histo hdrhistogram.Histogram
}

func (m *metrics) Reset() {
	// TODO(manik)
}

func (m *metrics) Namespaces() map[string]NamespaceMetrics {
	// TODO(manik)
	return nil
}

func (m *metrics) GlobalDefaultBucketMetrics() BucketMetrics {
	// TODO(manik)
	return nil
}

func New() Metrics {
	// TODO(manik): Proper init
	return &metrics{}
}

