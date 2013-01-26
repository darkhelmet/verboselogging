#!/usr/bin/env bash
set -e
export GOPATH=$PWD
go get -v verboselogging
(cd src && go test -v *)