#!/bin/sh
set -e

cat config.template.yaml | envsubst > config.yaml
./ch-p-proxy config.yaml
