package monitoring

import (
	"github.com/codahale/hdrhistogram"
)

type Monitoring struct {
	histo hdrhistogram.Histogram
}

func New() *Monitoring {
	// TODO(manik): Proper init
	return &Monitoring{}
}

