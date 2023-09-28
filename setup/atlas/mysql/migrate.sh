#!/usr/bin/env bash
set -euo pipefail
declare SELF
SELF=$(readlink -f "$0")
if [ -n "${SELF}" ]; then
  echo "Run $SELF"
fi
declare -r SELF_DIR=${SELF%/*}

declare -r MIG_DIR="${SELF_DIR:?}/migrations"
declare -r DB_TYPE='mysql'
declare -r DB_HOST_PORT="${DB_HOST_PORT:-localhost:3306}"
declare -r DB_USER_PASS="${DB_USER_PASS:-example:abcDe123}"

declare -r REV="example_rev"
declare -r MIG="file://${MIG_DIR:?}"
declare -r DEV="${DB_TYPE:?}://${DB_USER_PASS:?}@${DB_HOST_PORT:?}"
declare -r URL="${DB_TYPE:?}://${DB_USER_PASS:?}@${DB_HOST_PORT:?}/example"

mkdir -p "${MIG_DIR}"

echo '# atlas migrate status'
atlas migrate status --dir "${MIG}" --url "${DEV:?}" --revisions-schema "${REV}"

echo '# atlas schema clean'
atlas schema clean --url "${DEV:?}" --auto-approve

echo '# atlas migrate diff'
atlas migrate diff example --dir "${MIG}" --dev-url "${DEV:?}" --to "file://${SELF_DIR:?}/example.hcl"

echo '# atlas migrate validate'
atlas migrate validate --dir "${MIG}" --dev-url "${DEV:?}"

echo '# atlas migrate apply'
atlas migrate apply  --dir "${MIG}" --url "${DEV:?}" --revisions-schema "${REV}"

echo -e '\n----+----+----+----+----+----+----+----\n'

echo '# atlas migrate status'
atlas migrate status --dir "${MIG}" --url "${URL:?}" --revisions-schema "${REV}"

echo '# atlas migrate apply'
atlas migrate apply  --dir "${MIG}" --url "${URL:?}" --revisions-schema "${REV}"

#atlas schema inspect --url "${URL:?}"
