#!/usr/bin/env bash
set -e
export GOPATH=$PWD

function go_packages() {
    find . -iname '*.go' -exec dirname {} \; | sort | uniq | grep -v verboselogging
}

(
    cd src
    go_packages | while read pkg; do go test $pkg; done
)

go test -c verboselogging && ./verboselogging.test