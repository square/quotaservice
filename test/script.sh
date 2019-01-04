#!/bin/bash

set -ex

case "$TEST" in
  "UI")
    cd admin/public/
    npm install
    npm rebuild node-sass
    npm run lint
    npm run clean-dist
    npm run test
    ;;
  *)
    go vet $(go list ./... | grep -v /vendor/)
    go test -race -v ./...
    ;;
esac
