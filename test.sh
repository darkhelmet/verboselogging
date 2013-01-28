#!/usr/bin/env bash
set -e
export GOPATH=$PWD

IGNORE='verboselogging|gocheck'
go_packages() {
    find . -iname '*.go' -exec dirname {} \; | sort | uniq | grep -v -E $IGNORE
}

test_packages() {
    while read pkg; do go test $pkg; done
}

(
    cd src
    go_packages | test_packages
)

GOPATH=$PWD go test -c verboselogging && ./verboselogging.test