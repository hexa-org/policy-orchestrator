#!/usr/bin/env bash
set -euo pipefail
source ${REPO}/bin/support.sh

# ------------------------------------------------------------------------------
# Usage:
#
#   setup [options]
#
# Description:
#
#   Perform workspace/dev post-install setup tasks.
#
# Options:
#
#   -t, --target=<name>       Component to set up [default: all].
#                             One of:
#                               - all (default)
#                               - db
#   -h, --help                Print help text.
# ------------------------------------------------------------------------------

arg::parse "$@"

_main_() {
  case "${target}" in
    all )
      ensure::all
      ;;
    db )
      ensure::db
      ;;
    * )
      echo::fail "unknown setup target: '${target}'"
      ;;
  esac
}

ensure::all() {
  ensure::db
}

# db
# ----------------------------------------------------------
ensure::db() {
  check::db_user || setup::db_user
  check::db_data || setup::db_data
  exec::step "Running DB migrations" "migrate -verbose -path ${REPO}/databases/orchestrator -database 'postgres://orchestrator:orchestrator@localhost:5432/orchestrator_test?sslmode=disable' up"
}

check::db_user() {
  psql -c '\du' | awk '{print $1}' | grep --quiet '^orchestrator$'
}

setup::db_user() {
  createuser orchestrator
}

check::db_data() {
  psql -c '\list' | awk '{print $1}' | grep --quiet '^orchestrator_test$'
}

setup::db_data() {
  createdb orchestrator_test --owner orchestrator
  psql --quiet --command="alter user orchestrator with password 'orchestrator'"
  psql --quiet --command="grant all privileges on database orchestrator_test to orchestrator"
}

_main_ "$@"
