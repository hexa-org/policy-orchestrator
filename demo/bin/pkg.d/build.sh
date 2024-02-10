#!/usr/bin/env bash
set -euo pipefail
source ${REPO}/bin/support.sh

# ------------------------------------------------------------------------------
# Usage:
#
#   build [options]
#
# Description:
#
#   Build the Hexa image.
#
# Options:
#
#   -h, --help                Print help text.
# ------------------------------------------------------------------------------

arg::parse "$@"

_main_() {
  pushd ${REPO} > /dev/null
    exec::step "Building Hexa image" "pack build hexa --builder heroku/buildpacks:20"
  popd > /dev/null
}

_main_ "$@"
