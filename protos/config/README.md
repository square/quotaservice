# Configs

QuotaService configs are protobufs, so they can easily be serialized and persisted.

## Generating Go code from protos

Use `bin/compile_protos.sh` to compile protos, including config protos. However, since config
protos are also marshalled/unmarshalled to YAML, you then need to **hand-parse** the generated
`configs.pb.go` file (which will appear in this directory). For each of the generated structs
(`ServiceConfig`, `NamespaceConfig` and `BucketConfig`), you need to add a YAML name for each
field that contains **underscores**.

E.g.,

Generated:

```go
type BucketConfig struct {
	Name                string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Namespace           string `protobuf:"bytes,2,opt,name=namespace" json:"namespace,omitempty"`
	Size                int64  `protobuf:"varint,3,opt,name=size" json:"size,omitempty"`
	FillRate            int64  `protobuf:"varint,4,opt,name=fill_rate" json:"fill_rate,omitempty"`
	WaitTimeoutMillis   int64  `protobuf:"varint,5,opt,name=wait_timeout_millis" json:"wait_timeout_millis,omitempty"`
	MaxIdleMillis       int64  `protobuf:"varint,6,opt,name=max_idle_millis" json:"max_idle_millis,omitempty"`
	MaxDebtMillis       int64  `protobuf:"varint,7,opt,name=max_debt_millis" json:"max_debt_millis,omitempty"`
	MaxTokensPerRequest int64  `protobuf:"varint,8,opt,name=max_tokens_per_request" json:"max_tokens_per_request,omitempty"`
}
```

After updating:

```go
type BucketConfig struct {
	Name                string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Namespace           string `protobuf:"bytes,2,opt,name=namespace" json:"namespace,omitempty"`
	Size                int64  `protobuf:"varint,3,opt,name=size" json:"size,omitempty"`
	FillRate            int64  `protobuf:"varint,4,opt,name=fill_rate" json:"fill_rate,omitempty" yaml:"fill_rate"`
	WaitTimeoutMillis   int64  `protobuf:"varint,5,opt,name=wait_timeout_millis" json:"wait_timeout_millis,omitempty" yaml:"wait_timeout_millis"`
	MaxIdleMillis       int64  `protobuf:"varint,6,opt,name=max_idle_millis" json:"max_idle_millis,omitempty" yaml:"max_idle_millis"`
	MaxDebtMillis       int64  `protobuf:"varint,7,opt,name=max_debt_millis" json:"max_debt_millis,omitempty" yaml:"max_debt_millis"`
	MaxTokensPerRequest int64  `protobuf:"varint,8,opt,name=max_tokens_per_request" json:"max_tokens_per_request,omitempty" yaml:"max_tokens_per_request"`
}
```

Yes, I know this sucks.
