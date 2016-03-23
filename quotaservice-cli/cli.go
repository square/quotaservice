package main

import (
	"flag"
	"fmt"
	"os"
	"net/http"
	"io/ioutil"
)

const help = `Usage: quotaservice-cli -h host -p port (-n namespace) [COMMAND]

 where COMMAND is one of:

 -help:
 	Prints this help

 list:
 	Lists buckets in a given namespace. If -n is omitted, lists all namespaces and all buckets.

`

// TODO(manik) finish CLI
func main() {
	flag.Usage = func() {
		fmt.Println(help)
	}

	port := flag.Int("p", 8080, "Specify port to use.  Defaults to 8000.")
	host := flag.String("h", "localhost", "Specify host to use.  Defaults to localhost.")
	ns := flag.String("n", "", "Specify namespace.  Defaults to empty.")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Too few args!")
		flag.Usage()
		os.Exit(2)
	}

	switch flag.Args()[0] {
	case "list":
		list(*host, *port, *ns)
	default:
		flag.Usage()
		os.Exit(2)
	}
}

func list(host string, port int, ns string) {
	u := fmt.Sprintf("http://%s:%d/api/%s", host, port, ns)
	r, e := http.Get(u)
	if e != nil {
		fmt.Println("ERROR: ", e)
		return
	}

	defer r.Body.Close()
	b, e := ioutil.ReadAll(r.Body)
	if e != nil {
		fmt.Println("ERROR: ", e)
		return
	}

	fmt.Println("Config:")
	fmt.Println(string(b))
}
