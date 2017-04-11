// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	pb "github.com/maniksurtani/quotaservice/protos/config"
)

// Implements an interface for sorting server configs

type sortedConfigs []*pb.ServiceConfig

func (c sortedConfigs) Less(i, j int) bool {
	return c[i].Date > c[j].Date
}

func (c sortedConfigs) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c sortedConfigs) Len() int {
	return len(c)
}
