#!/bin/sh

# Licensed under the Apache License, Version 2.0
# Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

QS_HOME="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/.."

go test -v -bench=. -benchtime 30s -cpu=4 github.com/square/quotaservice/test/load -run ^$
