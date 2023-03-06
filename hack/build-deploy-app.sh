#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# shellcheck disable=SC1091
source hack/get-tools.sh

plantbuild push ./plantbuild/build.jsonnet

# https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#create-a-repository-dispatch-event

build_json=$(plantbuild show ./plantbuild/build.jsonnet | jq -r .services)
output_json='{"event_type": "deploy-test"}'
for key in $(echo  "$build_json" | jq -r keys[]); do
  app_name=$(echo "$key" | sed 's/build_image_//')
  image=$( echo "$build_json" | jq -r --arg key "$key" '.[$key].image')
  output_json=$(echo "$output_json" | jq -r \
      --arg app_name "$app_name" \
      --arg image "$image" \
      '.client_payload.github[$app_name].image = $image')
done

echo $output_json | jq .

curl -XPOST -H "Content-Type: application/json" \
    -H "Accept:  application/vnd.github.everest-preview+json" \
    -H "Authorization: Bearer $GITHUB_TOKEN" \
    https://api.github.com/repos/theplant/qor5-provisioning/dispatches \
    -d @- <<EOF
$output_json
EOF
