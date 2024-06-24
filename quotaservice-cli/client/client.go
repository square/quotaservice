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
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/square/quotaservice/config"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app   = kingpin.New("quotaservice-cli", "The quotaservice CLI tool.")
	debug = app.Flag("debug", "Print debug output").Default("false").Bool()
	env   = app.Flag("env", "Environment").Short('e').Enum("production", "staging", "local")

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
	addVersion   = add.Flag("version", "Current configuration version.").Short('v').Required().String()
	addNamespace = add.Arg("namespace", "Namespace to add to.").Required().String()
	addBucket    = add.Arg("bucket", "Bucket to add to.").String()

	// remove
	remove          = app.Command("remove", "Removes namespaces or buckets from a running configuration.")
	removeGDB       = remove.Flag("globaldefault", "Removes the global default bucket.").Short('g').Default("false").Bool()
	removeVersion   = remove.Flag("version", "Current configuration version.").Short('v').Required().String()
	removeNamespace = remove.Arg("namespace", "Namespace to remove.").Required().String()
	removeBucket    = remove.Arg("bucket", "Bucket to remove.").String()

	// update
	update          = app.Command("update", "Updates namespaces or buckets from a running configuration.")
	updateGDB       = update.Flag("globaldefault", "Updates the global default bucket.").Short('g').Default("false").Bool()
	updateFile      = update.Flag("file", "File from which to read configs.").Short('f').String()
	updateVersion   = update.Flag("version", "Current configuration version.").Short('v').Required().String()
	updateNamespace = update.Arg("namespace", "Namespace to update.").Required().String()
	updateBucket    = update.Arg("bucket", "Bucket to update.").String()
)

func RunClient(args []string) {
	switch kingpin.MustParse(app.Parse(args)) {
	// List
	case show.FullCommand():
		doShow(*showGDB, *showNamespace, *showBucket)
		break
	case add.FullCommand():
		doAdd(*addGDB, *addNamespace, *addBucket, *addVersion)
		break
	case remove.FullCommand():
		doRemove(*removeGDB, *removeNamespace, *removeBucket, *removeVersion)
		break
	case update.FullCommand():
		doUpdate(*updateGDB, *updateNamespace, *updateBucket, *updateVersion)
		break
	default:
		kingpin.FatalUsage("Unknown command; should never happen.")
	}
}

func doShow(gdb bool, namespace, bucket string) {
	validate(gdb, namespace, bucket, "")
	logf("Called show(gdb=%v, namespace=%v, bucket=%v)\n", gdb, namespace, bucket)
	url := createUrl(gdb, namespace, bucket)
	resp := connectToServer("GET", url, "")

	if *output == "" {
		fmt.Print(resp)
	} else {
		logf("Writing to %v\n", *output)
		f, err := os.Create(*output)
		kingpin.FatalIfError(err, "Cannot write to file %v", *output)
		_, err = f.WriteString(resp)
		kingpin.FatalIfError(err, "Cannot write to file %v", *output)
		err = f.Close()
		kingpin.FatalIfError(err, "Cannot write to file %v", *output)
	}
}

func doAdd(gdb bool, namespace, bucket, version string) {
	validate(gdb, namespace, bucket, version)
	logf("Called add(gdb=%v, namespace=%v, bucket=%v)\n", gdb, namespace, bucket)
	cfgBytes := readCfg(*addFile, namespace, bucket)
	url := createUrl(gdb, namespace, bucket)
	resp := connectToServer("POST", url, version, cfgBytes)
	fmt.Print(resp)
}

func doRemove(gdb bool, namespace, bucket, version string) {
	validate(gdb, namespace, bucket, version)
	logf("Called remove(gdb=%v, namespace=%v, bucket=%v)\n", gdb, namespace, bucket)
	url := createUrl(gdb, namespace, bucket)
	resp := connectToServer("DELETE", url, version)
	fmt.Print(resp)
}

func doUpdate(gdb bool, namespace, bucket, version string) {
	validate(gdb, namespace, bucket, version)
	logf("Called update(gdb=%v, namespace=%v, bucket=%v)\n", gdb, namespace, bucket)
	cfgBytes := readCfg(*updateFile, namespace, bucket)
	url := createUrl(gdb, namespace, bucket)
	resp := connectToServer("PUT", url, version, cfgBytes)
	fmt.Print(resp)
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

func validate(gdb bool, namespace, bucket, version string) {
	if version != "" {
		versionInt, err := strconv.Atoi(version)
		if err != nil || versionInt < 0 {
			kingpin.FatalUsage("Invalid version: %v", version)
		}
	}

	if gdb && (namespace != "" || bucket != "") {
		kingpin.FatalUsage("Bucket or namespace cannot be set if --globaldefault is used.")
	}

	if namespace == "" && bucket != "" {
		kingpin.FatalUsage("Namespace cannot be unset if bucket is set!")
	}
}

func connectToServer(method, url, version string, data ...[]byte) string {
	logf("Connecting (%v) to URL %v\n", method, url)

	var dataReader io.Reader

	switch len(data) {
	case 0:
		dataReader = nil
	case 1:
		dataReader = bytes.NewReader(data[0])
	default:
		panic("Should never get here")
	}

	cmdTokens := []string{"-sS", "-i", "-X", method, url, "-H", "Content-Type: application/json"}
	if dataReader != nil {
		cmdTokens = append(cmdTokens, "--data", string(data[0]))
	}

	if version != "" {
		cmdTokens = append(cmdTokens, "-H", fmt.Sprintf("Version: %v", version))
	}

	cmd := exec.Command("beyond-curl", cmdTokens...)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		kingpin.Fatalf("beyond-curl command failed with error: %v", err)
	}

	response := out.String()

	// Split the headers and body
	parts := strings.SplitN(response, "\r\n\r\n", 2)
	if len(parts) < 2 {
		kingpin.Fatalf("Unexpected response: %v\n", response)
	}
	headers := parts[0]
	body := parts[1]

	// Extract "Version" header
	headerLines := strings.Split(headers, "\r\n")
	for _, line := range headerLines {
		line = strings.ToLower(line)
		if strings.HasPrefix(line, "version:") {
			version := strings.TrimSpace(strings.TrimPrefix(line, "version:"))
			fmt.Printf("** Current version: %s\n", version)
			break
		}
	}

	return body
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

	host := "localhost:3000"
	if *env == "staging" {
		host = "quotaservice.stage.sqprod.co"
	} else if *env == "production" {
		host = "quotaservice.sqprod.co"
	}

	url := fmt.Sprintf("https://%v/api/%v", host, uri)
	return url
}

// logs to stdout if debug
func logf(format string, a ...interface{}) {
	if *debug {
		fmt.Printf(format, a...)
	}
}
