#!/usr/bin/env bash
set -euo pipefail
source ${REPO}/bin/support.sh

# ------------------------------------------------------------------------------
# Usage:
#
#   migrate create <desc>
#   migrate up
#   migrate down
#   migrate force <version>
#
# Description:
#
#   Utilities for working with golang-migrate
#
#   - "migrate create" will generate a new migration w/ project conventions.
#     e.g., "dev migrate create add_new_column"
#   - "migrate up" runs up migrations.
#   - "migrate down" runs down migrations.
#   - "migrate force" sets the migration version, ignoring dirty state.
#
# Options:
#
#   -h, --help                Print help text.
# ------------------------------------------------------------------------------

arg::parse "$@"

_main_() {
  pushd ${REPO} > /dev/null
    if [ "${create}" == "true" ] ; then
      migrate::create
    fi

    if [ "${up}" == "true" ] ; then
      migrate::up
    fi

    if [ "${down}" == "true" ] ; then
      migrate::down
    fi

    if [ "${force}" == "true" ] ; then
      migrate::force
    fi
  popd > /dev/null
}

db_url="postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable"

migrate::create() {
  migrate create -ext sql -dir databases/orchestrator -seq -digits 3 "${desc}"
}

migrate::up() {
  migrate -verbose \
    -path ${REPO}/databases/orchestrator \
    -database ${db_url} \
    up
}

migrate::down() {
  migrate -verbose \
    -path ${REPO}/databases/orchestrator \
    -database ${db_url} \
    down
}

migrate::force() {
  migrate -verbose \
    -path ${REPO}/databases/orchestrator \
    -database ${db_url} \
    force ${version}
}

_main_ "$@"
