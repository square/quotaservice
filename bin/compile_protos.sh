#!/bin/sh

QS_HOME="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/.."

protoc --go_out=plugins=grpc:. $QS_HOME/protos/*.proto --proto_path $QS_HOME
protoc --go_out=plugins=grpc:. $QS_HOME/protos/proto2/*.proto  --proto_path $QS_HOME
