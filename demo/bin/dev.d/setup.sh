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
#   Perform workspace development setup tasks.
#
# Options:
#
#   -t, --target=<name>       Component to set up [default: all].
#                             One of:
#                               - all (default)
#                               - db
#                               - opa
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
    opa )
      ensure::opa
      ;;
    * )
      echo::fail "unknown setup target: '${target}'"
      ;;
  esac
}

ensure::all() {
  ensure::db
  ensure::opa
}

# db
# ----------------------------------------------------------
ensure::db() {
  exec::stat "DB user" "check::db_user" "setup::db_user"
  exec::stat "DB data" "check::db_data" "setup::db_data"
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

# opa
# ----------------------------------------------------------
ensure::opa() {
  exec::stat "opa" "command -v opa" "install::opa"
}

install::opa() {
  pushd /tmp > /dev/null
    curl -L -o opa_darwin_amd64 https://openpolicyagent.org/downloads/v0.44.0/opa_darwin_amd64
    curl -L -o opa_darwin_amd64.sha256 https://openpolicyagent.org/downloads/v0.44.0/opa_darwin_amd64.sha256
    shasum -c opa_darwin_amd64.sha256
    chmod +x opa_darwin_amd64
    mv opa_darwin_amd64 "$(brew --prefix)/bin/opa2"
  popd > /dev/null
}

_main_ "$@"
