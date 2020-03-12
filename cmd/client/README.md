### Quotaservice CLI
This CLI makes HTTP calls into the quotaservice's admin REST endpoint which is also used by the quotaservice admin GUI.

The CLI makes use of [gopkg.in/alecthomas/kingpin.v2](https://godoc.org/gopkg.in/alecthomas/kingpin.v2).

#### Usage
Building and running the CLI with no options will present usage information:

```
$ quotaservice-cli
usage: quotaservice-cli [<flags>] <command> [<args> ...]

The quotaservice CLI tool.

Flags:
      --help              Show context-sensitive help (also try --help-long and --help-man).
  -v, --verbose           Verbose output
  -h, --host="localhost"  Host address
  -p, --port=80           Host port

Commands:
  help [<command>...]
    Show help.

  show [<flags>] [<namespace>] [<bucket>]
    Show configuration for the entire service, optionally filtered by namespace and/or bucket name.

  add [<flags>] [<namespace>] [<bucket>]
    Adds namespaces or buckets from a running configuration.

  remove [<flags>] [<namespace>] [<bucket>]
    Removes namespaces or buckets from a running configuration.

  update [<flags>] [<namespace>] [<bucket>]
    Updates namespaces or buckets from a running configuration.
```
