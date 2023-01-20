#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# shellcheck disable=SC1091
source hack/get-tools.sh

plantbuild push ./plantbuild/build.jsonnet
plantbuild k8s_set_images ./plantbuild/images.jsonnet

curl -XPOST -H "Content-Type: application/json" \
    -H "Accept:  application/vnd.github.everest-preview+json" \
    -H "Authorization: Bearer $GITHUB_TOKEN" \
    https://api.github.com/repos/theplant/qor5-provisioning/dispatches \
    ∙ -d '{"event_type":"deploy-test","client_payload":{"github":{"example":{"image":"public.ecr.aws/qor5/example:latest"}}}}'
