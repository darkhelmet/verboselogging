#!/usr/bin/env bash
set -e
export GOPATH=$PWD

function go_packages() {
    find . -iname '*.go' -exec dirname {} \; | sort | uniq
}

(
    cd src
    go_packages | while read pkg; do go test $pkg; done
)