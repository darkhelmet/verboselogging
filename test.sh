#!/usr/bin/env bash
set -e
export GOPATH=$PWD

IGNORE='verboselogging|gocheck'
function go_packages() {
    find . -iname '*.go' -exec dirname {} \; | sort | uniq | grep -v -E $IGNORE
}

(
    cd src
    go_packages | while read pkg; do go test $pkg; done
)

go test -c verboselogging && ./verboselogging.test