#!/usr/bin/env bash

set -euo pipefail

declare SELF
SELF=$(readlink -f "$0")
if [ -n "${SELF}" ]; then
  echo "Run $SELF"
fi

#declare -r  SELF_DIR=${SELF%/*}
#declare -r  UPPER_DIR=${SELF_DIR%/*}
#declare -rx GOPATH="${UPPER_DIR:?}/.go"

declare -rx GO111MODULE='on'

# Go version >= 1.18
go_install_bin() {
  local cmd_name=$1
  local version_tag=$2
  local src_path=$3
  local cmd_path="$GOPATH/bin/${cmd_name:?}"
  local package="${src_path:?}@${version_tag:?}"
  if [ ! -f "$cmd_path" ]; then
    go install "$package"
    echo "Installed OK: $package" '-->' "$(go version -v "$cmd_path")"
  else echo "Installed: $package" '-->' "$(go version -v "$cmd_path")"; fi
}

go_install_bin  sqlc v1.20.0  'github.com/sqlc-dev/sqlc/cmd/sqlc'
echo 'Installation finished'
