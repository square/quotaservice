# Configs

QuotaService configs are protobufs, so they can easily be serialized and persisted.

## Generating Go code from protos

Use `bin/compile_protos.sh` to compile protos, including config protos. However, since config
protos are also marshalled/unmarshalled to YAML and there is no first-class support, a custom
search-and-replace is executed. Please double-check that `configs.pb.go` has been properly created.
