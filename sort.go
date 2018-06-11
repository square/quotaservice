// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package quotaservice

import (
	"sort"

	"github.com/square/quotaservice/admin"
	pb "github.com/square/quotaservice/protos/config"
)

// Implements an interface for sorting server configs

type sortedConfigs []*admin.ConfigAndHash

func (c sortedConfigs) Less(i, j int) bool {
	return c[i].Date > c[j].Date
}

func (c sortedConfigs) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c sortedConfigs) Len() int {
	return len(c)
}

func addHashesAndSortDesc(cfgs []*pb.ServiceConfig) []*admin.ConfigAndHash {
	sorted := make(sortedConfigs, 0, len(cfgs))
	for _, cfg := range cfgs {
		sorted = append(sorted, admin.NewConfigAndHash(cfg))
	}
	sort.Sort(sorted)

	return sorted
}
