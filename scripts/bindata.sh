#!/bin/sh

set -e

cd "$(dirname "$0")/.."

go-bindata -o internal/bindata/bindata.go -pkg bindata -prefix assets assets
