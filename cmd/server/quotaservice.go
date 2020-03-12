// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package main

import (
	"os"

	"github.com/square/quotaservice/cmd/server/server"
	"github.com/square/quotaservice/config"
)

func main() {
	cfg := config.NewDefaultServiceConfig()
	server.RunServer(cfg, os.Args[1:])
}
