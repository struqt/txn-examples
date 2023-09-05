#!/usr/bin/env bash

set -euo pipefail

declare SELF
SELF=$(readlink -f "$0")
if [ -n "${SELF}" ]; then
  echo "Run $SELF"
fi
declare -r SELF_DIR=${SELF%/*}
declare -r  OUT_DIR=${SELF_DIR:?}/build

build_release() {
  local module="$1"
  local file
  file="${module:?}_demo_$(go env GOHOSTOS)_$(go env GOARCH)"
  pushd "${SELF_DIR:?}/${module:?}"
  go mod tidy
  go get -d -v -u all
  go mod tidy
  gofmt -w -l -d -s .
  go build -ldflags "-s -w" -o "${OUT_DIR:?}/${file:?}"
  echo " Built: ${OUT_DIR:?}/${file:?}"
  popd
}

which go

mkdir -p  "${OUT_DIR:?}"
rm    -rf "${OUT_DIR:?}"/*

build_release mysql
build_release pg
build_release pgx
