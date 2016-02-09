#!/bin/sh

protoc --go_out=plugins=grpc:. ./protos/*.proto --proto_path ./
