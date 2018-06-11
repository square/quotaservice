// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package quotaservice

import (
	"testing"

	"github.com/square/quotaservice/config"
	pb "github.com/square/quotaservice/protos/config"
)

func TestSortingAndHashes(t *testing.T) {
	cfgs := make([]*pb.ServiceConfig, 3)
	cfgs[0] = config.NewDefaultServiceConfig()
	cfgs[1] = config.NewDefaultServiceConfig()
	cfgs[2] = config.NewDefaultServiceConfig()

	cfgs[0].Date = 100
	cfgs[1].Date = 90
	cfgs[2].Date = 110

	reverseSorted := addHashesAndSortDesc(cfgs)
	expectedDates := []int64{110, 100, 90}

	for i, c := range reverseSorted {
		if c.Date != expectedDates[i] {
			t.Errorf("Expected Date on element %v to be %v but was %v", i, expectedDates[i], c.Date)
		}

		if c.Hash == "" {
			t.Errorf("Expected non-empty Hash on element %v", i)
		}
	}
}
