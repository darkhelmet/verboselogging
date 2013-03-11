#!/usr/bin/env bash
set -e
export GOPATH=$PWD

IGNORE='verboselogging|gocheck'
go-packages() {
    find . -iname '*.go' -exec dirname {} \; | sort | uniq | grep -v -E $IGNORE
}

test-packages() {
    while read pkg; do go test $pkg; done
}

(
    cd src
    go-packages | test-packages
)

GOPATH=$PWD go test -c verboselogging && ./verboselogging.test
