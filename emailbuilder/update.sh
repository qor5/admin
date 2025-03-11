#!/usr/bin/env bash

set -e

curl -L https://codeload.github.com/qor5/email-builder/zip/refs/heads/release?token= -o email-builder-spa.zip

rm -rf dist

unzip email-builder-spa.zip

mv email-builder-release dist

rm email-builder-spa.zip

echo "Done! The email-builder-spa has been updated."
