#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail


curl -L https://raw.githubusercontent.com/theplant/plantbuild/master/plantbuild > ./pb \
  && chmod +x ./pb \
  && mv ./pb /usr/local/bin/plantbuild

curl -LO https://storage.googleapis.com/kubernetes-release/release/"$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)"/bin/linux/amd64/kubectl \
  && chmod +x ./kubectl \
  && mv ./kubectl /usr/local/bin/kubectl

kubectl version --client
