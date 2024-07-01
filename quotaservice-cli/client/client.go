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

	"github.com/alecthomas/kingpin/v2"

	"github.com/square/quotaservice/config"
)

type QuotaserviceClient struct {
	client  *http.Client
	verbose bool
	host    string
	port    string
}

func NewQuotaserviceClient(client *http.Client, verbose bool, host string, port int) *QuotaserviceClient {

	return &QuotaserviceClient{
		client:  client,
		verbose: verbose,
		host:    host,
		port:    fmt.Sprintf("%v", port),
	}
}

func (c *QuotaserviceClient) DoShow(gdb bool, namespace, bucket, output string) {
	c.validate(gdb, namespace, bucket)
	c.logf("Called show(gdb=%v, namespace=%v, bucket=%v)\n", gdb, namespace, bucket)
	url := c.createUrl(gdb, namespace, bucket)
	resp := c.connectToServer("GET", url)
	defer func() { _ = resp.Body.Close() }()
	body, e := ioutil.ReadAll(resp.Body)
	kingpin.FatalIfError(e, "Error reading HTTP response")

	if output == "" {
		fmt.Print(string(body))
	} else {
		c.logf("Writing to %v\n", output)
		f, err := os.Create(output)
		kingpin.FatalIfError(err, "Cannot write to file %v", output)
		_, err = f.WriteString(string(body))
		kingpin.FatalIfError(err, "Cannot write to file %v", output)
		err = f.Close()
		kingpin.FatalIfError(err, "Cannot write to file %v", output)
	}
}

func (c *QuotaserviceClient) DoAdd(gdb bool, namespace, bucket, file string) {
	c.validate(gdb, namespace, bucket)
	c.logf("Called add(gdb=%v, namespace=%v, bucket=%v)\n", gdb, namespace, bucket)
	cfgBytes := c.readCfg(file, namespace, bucket)
	url := c.createUrl(gdb, namespace, bucket)
	resp := c.connectToServer("POST", url, cfgBytes)
	_ = resp.Body.Close()
}

func (c *QuotaserviceClient) DoRemove(gdb bool, namespace, bucket string) {
	c.validate(gdb, namespace, bucket)
	c.logf("Called remove(gdb=%v, namespace=%v, bucket=%v)\n", gdb, namespace, bucket)
	url := c.createUrl(gdb, namespace, bucket)
	resp := c.connectToServer("DELETE", url)
	_ = resp.Body.Close()
}

func (c *QuotaserviceClient) DoUpdate(gdb bool, namespace, bucket, file string) {
	c.validate(gdb, namespace, bucket)
	c.logf("Called update(gdb=%v, namespace=%v, bucket=%v)\n", gdb, namespace, bucket)
	cfgBytes := c.readCfg(file, namespace, bucket)
	url := c.createUrl(gdb, namespace, bucket)
	resp := c.connectToServer("PUT", url, cfgBytes)
	_ = resp.Body.Close()
}

func (c *QuotaserviceClient) readCfg(f, namespace, bucket string) []byte {
	var cfgBytes []byte
	var e error

	if f == "" {
		f = "STDIN"
		fmt.Print("Please input the config. Press ctrl-D to continue.")
		cfgBytes, e = ioutil.ReadAll(os.Stdin)
	} else {
		cfgBytes, e = ioutil.ReadFile(f)
	}

	kingpin.FatalIfError(e, "Could not read config from %v", f)
	c.logf("Read config %v from %v\n", string(cfgBytes), f)
	c.validateJSON(cfgBytes, namespace, bucket)
	return cfgBytes
}

func (c *QuotaserviceClient) validateJSON(j []byte, namespace, bucket string) {
	var js map[string]interface{}
	if json.Unmarshal(j, &js) != nil {
		kingpin.Fatalf("Config read isn't valid JSON!\n")
	}

	if namespace == "" {
		// Global default bucket
		c.checkField("name", config.DefaultBucketName, js, func(val string) {
			kingpin.Fatalf("Global default bucket cannot have name '%v'", val)
		})
		c.checkField("namespace", config.GlobalNamespace, js, func(val string) {
			kingpin.Fatalf("Global default bucket cannot have namespace '%v'", val)
		})
	} else if bucket == "" {
		// We're just updating a namespace.
		c.checkField("name", namespace, js, func(val string) {
			kingpin.Fatalf("Attempting to configure namespace '%v' but config provided is for '%v'", namespace, val)
		})
	} else {
		c.checkField("name", bucket, js, func(val string) {
			kingpin.Fatalf("Attempting to configure bucket '%v' but config provided is for '%v'", bucket, val)
		})
		c.checkField("namespace", namespace, js, func(val string) {
			kingpin.Fatalf("Attempting to configure namespace '%v' but config provided is for '%v'", namespace, val)
		})
	}
}

func (c *QuotaserviceClient) checkField(field, expected string, js map[string]interface{}, errHandler func(string)) {
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

func (c *QuotaserviceClient) validate(gdb bool, namespace, bucket string) {
	if gdb && (namespace != "" || bucket != "") {
		kingpin.FatalUsage("Bucket or namespace cannot be set if --globaldefault is used.")
	}

	if namespace == "" && bucket != "" {
		kingpin.FatalUsage("Namespace cannot be unset if bucket is set!")
	}
}

func (c *QuotaserviceClient) connectToServer(method, url string, data ...[]byte) *http.Response {
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

	resp, e := c.client.Do(r)
	kingpin.FatalIfError(e, "HTTP error")

	c.logf("Response Status: %v\n", resp.Status)
	c.logf("Response Headers: %v\n", resp.Header)
	if resp.StatusCode != 200 {
		defer func() { _ = resp.Body.Close() }()
		body, _ := ioutil.ReadAll(resp.Body)
		kingpin.Fatalf("HTTP request failed with status %v and reason %v\n", resp.StatusCode, string(body))
	}

	return resp
}

func (c *QuotaserviceClient) createUrl(gdb bool, namespace, bucket string) string {
	uri := ""
	if !gdb {
		if namespace != "" {
			uri = namespace
		}

		if bucket != "" {
			uri += "/" + bucket
		}
	}

	url := fmt.Sprintf("https://%v:%v/api/%v", c.host, c.port, uri)
	c.logf("Connecting to URL %v\n", url)
	return url
}

// logs to stdout if verbose
func (c *QuotaserviceClient) logf(format string, a ...interface{}) {
	if c.verbose {
		fmt.Printf(format, a...)
	}
}
