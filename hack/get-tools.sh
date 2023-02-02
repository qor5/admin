#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail


curl -L https://raw.githubusercontent.com/theplant/plantbuild/master/plantbuild > ./pb \
  && chmod +x ./pb \
  && mv ./pb /usr/local/bin/plantbuild


