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
  exec::step "Building Hexa image" "pack build hexa --builder heroku/buildpacks:20"
}

_main_ "$@"
