#!/bin/sh

# Licensed under the Apache License, Version 2.0
# Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

QS_HOME="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/.."

go test -v -bench=. -benchtime 240s -cpu=8 github.com/maniksurtani/quotaservice/test/load -run ^$
