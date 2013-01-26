#!/usr/bin/env bash
set -e
export GOPATH=$PWD
go get -v verboselogging
go get -v launchpad.net/gocheck