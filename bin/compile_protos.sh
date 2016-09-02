#!/bin/sh

# Licensed under the Apache License, Version 2.0
# Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

set -e

protoc --go_out=plugins=grpc:. ./protos/*.proto --proto_path ./
protoc --go_out=plugins=grpc:. ./protos/config/*.proto --proto_path ./

# need the .2 extension so that this works on os x and linux equally
sed -i.2 -e 's/\(json:"\([^,]*\),omitempty"\)/\1 yaml:"\2"/' ./protos/config/configs.pb.go
rm ./protos/config/configs.pb.go.2

echo "Protos compiled. If you made any changes to protos/config/configs.proto, then please read protos/config/README.md now."
