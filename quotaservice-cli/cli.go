// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

// Package implements a CLI for administering the quotaservice.
package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	app     = kingpin.New("quotaservice-cli", "The quotaservice CLI tool.")
	verbose = app.Flag("verbose", "Verbose output").Short('v').Default("false").Bool()
	host    = app.Flag("host", "Host address").Short('h').Default("localhost").String()
	port    = app.Flag("port", "Host port").Short('p').Default("80").Int()

	// show
	show          = app.Command("show", "Show configuration for the entire service, optionally filtered by namespace and/or bucket name.")
	showGDB       = show.Flag("globaldefault", "Only show configs for the global default bucket.").Short('g').Default("false").Bool()
	showNamespace = show.Arg("namespace", "Only show configs for a given namespace.").String()
	showBucket    = show.Arg("bucket", "Only show configs for a given bucket in a given namespace.").String()

	// add
	add          = app.Command("add", "Adds namespaces or buckets from a running configuration.")
	addGDB       = add.Flag("globaldefault", "Apply to the global default bucket.").Short('g').Default("false").Bool()
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
	updateNamespace = update.Arg("namespace", "Namespace to update.").String()
	updateBucket    = update.Arg("bucket", "Bucket to update.").String()
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	// List
	case show.FullCommand():
		doShow(*showGDB, *showNamespace, *showBucket)
		break
	case add.FullCommand():
		doAdd(*addGDB, *addNamespace, *addBucket)
		break
	case remove.FullCommand():
		doRemove(*removeGDB, *removeNamespace, *removeBucket)
		break
	case update.FullCommand():
		doRemove(*updateGDB, *updateNamespace, *updateBucket)
		break
	default:
		kingpin.FatalUsage("Unknown command; should never happen.")
	}
}

func doShow(gdb bool, namespace, bucket string) {
	validate(gdb, namespace, bucket)
	logf("Called show(%v, %v, %v)\n", gdb, namespace, bucket)
	url := createUrl(gdb, namespace, bucket)
	resp := connectToServer("GET", url)
	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	kingpin.FatalIfError(e, "Error reading HTTP response")
	fmt.Println(string(body))
}

func doAdd(gdb bool, namespace, bucket string) {
	validate(gdb, namespace, bucket)
	logf("Called add(%v, %v, %v)\n", gdb, namespace, bucket)
	cfgBytes, e := ioutil.ReadAll(os.Stdin)
	kingpin.FatalIfError(e, "Could not read config from stdin")
	cfg := string(cfgBytes)
	logf("Read config %v from stdin\n", cfg)
	_ = createUrl(gdb, namespace, bucket)
	// TODO: Make REST call to server to add config and print status
}

func doRemove(gdb bool, namespace, bucket string) {
	validate(gdb, namespace, bucket)
	logf("Called remove(%v, %v, %v)\n", gdb, namespace, bucket)
	_ = createUrl(gdb, namespace, bucket)
	// TODO: Make REST call to server to remove config and print status
}

func doUpdate(gdb bool, namespace, bucket string) {
	validate(gdb, namespace, bucket)
	logf("Called update(%v, %v, %v)\n", gdb, namespace, bucket)
	_ = createUrl(gdb, namespace, bucket)
	// TODO: Make REST call to server to update config and print status
}

func validate(gdb bool, namespace, bucket string) {
	if gdb && (namespace != "" || bucket != "") {
		kingpin.FatalUsage("Bucket or namespace cannot be set if --globaldefault is used.")
	}

	if namespace == "" && bucket != "" {
		kingpin.FatalUsage("Namespace cannot be unset if bucket is set!")
	}
}

func connectToServer(method, url string) *http.Response {
	r, e := http.NewRequest(method, url, nil)
	kingpin.FatalIfError(e, "HTTP error")

	client := &http.Client{}
	resp, e := client.Do(r)
	kingpin.FatalIfError(e, "HTTP error")

	logf("Response Status: %v\n", resp.Status)
	logf("Response Headers: %v\n", resp.Header)
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		kingpin.Fatalf("HTTP request failed with status %v and reason %v\n", resp.StatusCode, string(body))
	}

	return resp
}

func createUrl(gdb bool, namespace, bucket string) string {
	uri := ""
	if !gdb {
		if namespace != "" {
			uri = namespace
		}

		if bucket != "" {
			uri += "/" + bucket
		}
	}

	url := fmt.Sprintf("http://%v:%v/api/%v", *host, *port, uri)
	logf("Connecting to URL %v\n", url)
	return url
}

// logs to stdout if verbose
func logf(format string, a ...interface{}) {
	if *verbose {
		fmt.Printf(format, a...)
	}
}
