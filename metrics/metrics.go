package metrics

import (
	"github.com/codahale/hdrhistogram"
)

type Metrics interface {
	// TODO(manik) What are these interface methods?
	TODO()
}

type metrics struct {
	histo hdrhistogram.Histogram
}

func (m *metrics) TODO() {}

func New() Metrics {
	// TODO(manik): Proper init
	return &metrics{}
}

