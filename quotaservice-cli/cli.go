// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

// Package implements a CLI for administering the quotaservice.
package main

import (
	"net/http"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/square/quotaservice/quotaservice-cli/client"
)

var (
	app     = kingpin.New("quotaservice-cli", "The quotaservice CLI tool.")
	verbose = app.Flag("verbose", "Verbose output").Short('v').Default("false").Bool()
	host    = app.Flag("host", "Host address").Short('h').Default("localhost").String()
	port    = app.Flag("port", "Host port").Short('p').Default("443").Int()

	// show
	show          = app.Command("show", "Show configuration for the entire service, optionally filtered by namespace and/or bucket name.")
	showGDB       = show.Flag("globaldefault", "Only show configs for the global default bucket.").Short('g').Default("false").Bool()
	output        = show.Flag("out", "Send output to file.").Short('o').String()
	showNamespace = show.Arg("namespace", "Only show configs for a given namespace.").String()
	showBucket    = show.Arg("bucket", "Only show configs for a given bucket in a given namespace.").String()

	// add
	add          = app.Command("add", "Adds namespaces or buckets from a running configuration.")
	addGDB       = add.Flag("globaldefault", "Apply to the global default bucket.").Short('g').Default("false").Bool()
	addFile      = add.Flag("file", "File from which to read configs.").Short('f').String()
	addNamespace = add.Arg("namespace", "Namespace to add to.").String()
	addBucket    = add.Arg("bucket", "Bucket to add to.").String()

	// remove
	remove          = app.Command("remove", "Removes namespaces or buckets from a running configuration.")
	removeGDB       = remove.Flag("globaldefault", "Removes the global default bucket.").Short('g').Default("false").Bool()
	removeNamespace = remove.Arg("namespace", "Namespace to remove.").String()
	removeBucket    = remove.Arg("bucket", "Bucket to remove.").String()

	// update
	update          = app.Command("update", "Updates namespaces or buckets from a running configuration.")
	updateGDB       = update.Flag("globaldefault", "Updates the global default bucket.").Short('g').Default("false").Bool()
	updateFile      = update.Flag("file", "File from which to read configs.").Short('f').String()
	updateNamespace = update.Arg("namespace", "Namespace to update.").String()
	updateBucket    = update.Arg("bucket", "Bucket to update.").String()
)

func RunClient(args []string) {
	cmd, err := app.Parse(args)
	c := client.NewQuotaserviceClient(&http.Client{}, *verbose, *host, *port)
	switch kingpin.MustParse(cmd, err) {
	case show.FullCommand():
		c.DoShow(*showGDB, *showNamespace, *showBucket, *output)
		break
	case add.FullCommand():
		c.DoAdd(*addGDB, *addNamespace, *addBucket, *addFile)
		break
	case remove.FullCommand():
		c.DoRemove(*removeGDB, *removeNamespace, *removeBucket)
		break
	case update.FullCommand():
		c.DoUpdate(*updateGDB, *updateNamespace, *updateBucket, *updateFile)
		break
	default:
		kingpin.FatalUsage("Unknown command; should never happen.")
	}
}

func main() {
	RunClient(os.Args[1:])
}
