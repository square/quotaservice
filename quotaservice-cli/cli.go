package main

import (
	"flag"
	"fmt"
	"os"
)

const help = `Usage: quotaservice-cli -h host -p port (-n namespace) [COMMAND]

 where COMMAND is one of:

 -help:
 	Prints this help

 list:
 	Lists buckets in a given namespace. If -n is omitted, lists all namespaces and all buckets.

`

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
		fmt.Println("Listing... ", host, port, ns)
	default:
		flag.Usage()
		os.Exit(2)
	}
}
