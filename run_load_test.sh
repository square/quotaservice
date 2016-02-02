#!/bin/bash

go test -v -bench=. -benchtime 240s -cpu=8 github.com/maniksurtani/qs/quotaservice/server/loadtest -run ^$
