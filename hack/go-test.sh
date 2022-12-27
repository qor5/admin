#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

source go-test.env
go test -p=1 -count=1 ./...
