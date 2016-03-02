#!/bin/sh

# Licensed under the Apache License, Version 2.0
# Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

QS_HOME="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/.."

go run $QS_HOME/test/service/main.go
