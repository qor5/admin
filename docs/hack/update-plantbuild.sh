#!/usr/bin/env sh

set -o errexit
set -o nounset
set -o pipefail

find . -name '*.jsonnet' -exec jsonnetfmt -i '{}' +
