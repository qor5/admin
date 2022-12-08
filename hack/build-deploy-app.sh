#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# shellcheck disable=SC1091
source hack/get-tools.sh

plantbuild push ./plantbuild/build.jsonnet
plantbuild k8s_set_images ./plantbuild/images.jsonnet

NAMESPACE="qor5-test"
echo "kubectl -n $NAMESPACE get deploy -o name | xargs -n1 kubectl -n $NAMESPACE rollout status --timeout 150s" | $KUBECTL_BASH
