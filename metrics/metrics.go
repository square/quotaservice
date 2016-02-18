package metrics

import (
	"github.com/codahale/hdrhistogram"
)

type Metrics struct {
	histo hdrhistogram.Histogram
}

func New() *Metrics {
	// TODO(manik): Proper init
	return &Metrics{}
}

