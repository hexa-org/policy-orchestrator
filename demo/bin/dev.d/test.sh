#!/usr/bin/env bash
set -euo pipefail
source ${REPO}/bin/support.sh

# ------------------------------------------------------------------------------
# Usage:
#
#   test [options]
#
# Description:
#
#   Run the test suite and related tasks.
#
# Options:
#
#   -c, --clean               Clear the test cache first.
#   -h, --help                Print help text.
# ------------------------------------------------------------------------------

arg::parse "$@"

_main_() {
  pushd ${REPO} > /dev/null
    if [ "${clean}" == "true" ] ; then
      go clean -testcache
    fi

    go test ./...
  popd > /dev/null
}

_main_ "$@"
