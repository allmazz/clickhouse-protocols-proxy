#!/bin/sh
set -e

cat config.template.yaml | envsubst > config.yaml
./clickhouse-protocol-proxy config.yaml
