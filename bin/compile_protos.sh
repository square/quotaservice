#!/bin/sh

# Licensed under the Apache License, Version 2.0
# Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

protoc --go_out=plugins=grpc:. ./protos/*.proto --proto_path ./
protoc --go_out=plugins=grpc:. ./protos/config/*.proto --proto_path ./
