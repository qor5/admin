#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# shellcheck disable=SC1091
#source hack/get-tools.sh

#plantbuild push ./plantbuild/build.jsonnet

# https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#create-a-repository-dispatch-event
#
images=$(pb show ./plantbuild/build.jsonnet | jq -r .services[].image)
set +e # ignore error so one failure doesn't stop the whole thing

for image in $images; do
  echo "deploying $image"

  curl -XPOST -H "Content-Type: application/json" \
      -H "Accept:  application/vnd.github.everest-preview+json" \
      -H "Authorization: Bearer $GITHUB_TOKEN" \
      https://api.github.com/repos/theplant/qor5-provisioning/dispatches \
      -d '{"event_type":"deploy-test","client_payload":{"github":{"example":{"image":"'"$image"'"}}}}'
done

