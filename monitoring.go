package quotaservice

import (
	"github.com/codahale/hdrhistogram"
)

type Monitoring struct {
	histo hdrhistogram.Histogram
}

func newMonitoring() *Monitoring {
	// TODO(manik): Proper init
	return &Monitoring{}
}

