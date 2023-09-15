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
  echo "Build: ${OUT_DIR:?}/${file:?}"
  pushd "${SELF_DIR:?}/${module:?}" >/dev/null || exit 1
  go mod tidy
  go get -d -v -u all
  go mod tidy
  gofmt -w -l -d -s .
  go build -ldflags "-s -w" -o "${OUT_DIR:?}/${file:?}"
  echo -e "Built: ${OUT_DIR:?}/${file:?}\n"
  popd >/dev/null || exit 1
}

which go

mkdir -p  "${OUT_DIR:?}"
rm    -rf "${OUT_DIR:?}"/*

build_release mongo
build_release sqlc/mysql
build_release sqlc/pg
build_release sqlc/pgx
build_release sqlc/sqlite
