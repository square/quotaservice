// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

// Package implements a CLI for administering the quotaservice.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/square/quotaservice/config"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app     = kingpin.New("quotaservice-cli", "The quotaservice CLI tool.")
	verbose = app.Flag("verbose", "Verbose output").Short('v').Default("false").Bool()
	host    = app.Flag("host", "Host address").Short('h').Default("localhost").String()
	port    = app.Flag("port", "Host port").Short('p').Default("80").Int()

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
	switch kingpin.MustParse(app.Parse(args)) {
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
		doUpdate(*updateGDB, *updateNamespace, *updateBucket)
		break
	default:
		kingpin.FatalUsage("Unknown command; should never happen.")
	}
}

func doShow(gdb bool, namespace, bucket string) {
	validate(gdb, namespace, bucket)
	logf("Called show(gdb=%v, namespace=%v, bucket=%v)\n", gdb, namespace, bucket)
	url := createUrl(gdb, namespace, bucket)
	resp := connectToServer("GET", url)
	defer func() { _ = resp.Body.Close() }()
	body, e := ioutil.ReadAll(resp.Body)
	kingpin.FatalIfError(e, "Error reading HTTP response")

	if *output == "" {
		fmt.Print(string(body))
	} else {
		logf("Writing to %v\n", *output)
		f, err := os.Create(*output)
		kingpin.FatalIfError(err, "Cannot write to file %v", *output)
		_, err = f.WriteString(string(body))
		kingpin.FatalIfError(err, "Cannot write to file %v", *output)
		err = f.Close()
		kingpin.FatalIfError(err, "Cannot write to file %v", *output)
	}
}

func doAdd(gdb bool, namespace, bucket string) {
	validate(gdb, namespace, bucket)
	logf("Called add(gdb=%v, namespace=%v, bucket=%v)\n", gdb, namespace, bucket)
	cfgBytes := readCfg(*addFile, namespace, bucket)
	url := createUrl(gdb, namespace, bucket)
	resp := connectToServer("POST", url, cfgBytes)
	_ = resp.Body.Close()
}

func doRemove(gdb bool, namespace, bucket string) {
	validate(gdb, namespace, bucket)
	logf("Called remove(gdb=%v, namespace=%v, bucket=%v)\n", gdb, namespace, bucket)
	url := createUrl(gdb, namespace, bucket)
	resp := connectToServer("DELETE", url)
	_ = resp.Body.Close()
}

func doUpdate(gdb bool, namespace, bucket string) {
	validate(gdb, namespace, bucket)
	logf("Called update(gdb=%v, namespace=%v, bucket=%v)\n", gdb, namespace, bucket)
	cfgBytes := readCfg(*updateFile, namespace, bucket)
	url := createUrl(gdb, namespace, bucket)
	resp := connectToServer("PUT", url, cfgBytes)
	_ = resp.Body.Close()
}

func readCfg(f, namespace, bucket string) []byte {
	var cfgBytes []byte
	var e error

	if f == "" {
		f = "STDIN"
		cfgBytes, e = ioutil.ReadAll(os.Stdin)
	} else {
		cfgBytes, e = ioutil.ReadFile(f)
	}

	kingpin.FatalIfError(e, "Could not read config from %v", f)
	logf("Read config %v from %v\n", string(cfgBytes), f)
	validateJSON(cfgBytes, namespace, bucket)
	return cfgBytes
}

func validateJSON(j []byte, namespace, bucket string) {
	var js map[string]interface{}
	if json.Unmarshal(j, &js) != nil {
		kingpin.Fatalf("Config read isn't valid JSON!\n")
	}

	if namespace == "" {
		// Global default bucket
		checkField("name", config.DefaultBucketName, js, func(val string) {
			kingpin.Fatalf("Global default bucket cannot have name '%v'", val)
		})
		checkField("namespace", config.GlobalNamespace, js, func(val string) {
			kingpin.Fatalf("Global default bucket cannot have namespace '%v'", val)
		})
	} else if bucket == "" {
		// We're just updating a namespace.
		checkField("name", namespace, js, func(val string) {
			kingpin.Fatalf("Attempting to configure namespace '%v' but config provided is for '%v'", namespace, val)
		})
	} else {
		checkField("name", bucket, js, func(val string) {
			kingpin.Fatalf("Attempting to configure bucket '%v' but config provided is for '%v'", bucket, val)
		})
		checkField("namespace", namespace, js, func(val string) {
			kingpin.Fatalf("Attempting to configure namespace '%v' but config provided is for '%v'", namespace, val)
		})
	}
}

func checkField(field, expected string, js map[string]interface{}, errHandler func(string)) {
	if val, exists := js[field]; exists {
		if val != expected && val != "" {
			vStr, ok := val.(string)
			if !ok {
				vStr = "[NOT A STRING]"
			}
			errHandler(vStr)
		}
	}
}

func validate(gdb bool, namespace, bucket string) {
	if gdb && (namespace != "" || bucket != "") {
		kingpin.FatalUsage("Bucket or namespace cannot be set if --globaldefault is used.")
	}

	if namespace == "" && bucket != "" {
		kingpin.FatalUsage("Namespace cannot be unset if bucket is set!")
	}
}

func connectToServer(method, url string, data ...[]byte) *http.Response {
	var dataReader io.Reader

	switch len(data) {
	case 0:
		dataReader = nil
	case 1:
		dataReader = bytes.NewReader(data[0])
	default:
		panic("Should never get here")
	}

	r, e := http.NewRequest(method, url, dataReader)
	kingpin.FatalIfError(e, "HTTP error")

	client := &http.Client{}
	resp, e := client.Do(r)
	kingpin.FatalIfError(e, "HTTP error")

	logf("Response Status: %v\n", resp.Status)
	logf("Response Headers: %v\n", resp.Header)
	if resp.StatusCode != 200 {
		defer func() { _ = resp.Body.Close() }()
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
