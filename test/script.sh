#!/bin/bash

set -ex

case "$TEST" in
  "UI")
    cd admin/public/
    npm install
    npm run lint
    npm run dist
    ;;
  *)
    go test -race -v ./...
    ;;
esac
