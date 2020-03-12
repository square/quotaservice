// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

// Package implements a CLI for administering the quotaservice.
package main

import (
	"os"

	"github.com/square/quotaservice/cmd/client/client"
)

func main() {
	client.RunClient(os.Args[1:])
}
