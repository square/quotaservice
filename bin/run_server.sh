#!/bin/sh

QS_HOME="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/.."

go run $QS_HOME/test/service/main.go
