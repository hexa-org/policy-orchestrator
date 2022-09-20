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
#   Builds the Hexa Policy Orchestrator.
#
# Options:
#
#   -h, --help      Print help text.
# ------------------------------------------------------------------------------

arg::parse "$@"

_main_() {
  exec::step "Building Policy Orchestrator" "pack build hexa --builder heroku/buildpacks:20"
}

_main_ "$@"
