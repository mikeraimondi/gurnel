#!/bin/sh

set -e

cd "$(dirname "$0")/.."

go-bindata -o internal/gurnel/bindata.go -pkg gurnel -prefix assets -debug assets
