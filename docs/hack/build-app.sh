#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# shellcheck disable=SC1091
source hack/get-tools.sh

plantbuild push ./plantbuild/build.jsonnet -l false

echo "now please MANUALLY fill the image tag in the provisioning repo..."
