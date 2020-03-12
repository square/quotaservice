// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package main

import (
	"github.com/square/quotaservice/config"
	"github.com/square/quotaservice/cmd/server/server"
)

func main() {
	cfg := config.NewDefaultServiceConfig()
	server.RunServer(cfg)
}
