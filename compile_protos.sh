#!/bin/bash
protoc --go_out=plugins=grpc:. ./quotaservice/protos/*.proto
protoc --go_out=plugins=grpc:. ./quotaservice/protos/proto2/*.proto
