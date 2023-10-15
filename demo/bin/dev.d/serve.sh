#!/usr/bin/env bash
set -euo pipefail
source ${REPO}/bin/support.sh

# ------------------------------------------------------------------------------
# Usage:
#
#   serve
#   serve demo
#   serve hexa
#
# Description:
#
#   Run the Hexa applications (in development).
#
# Options:
#
#   -h, --help                Print help text.
# ------------------------------------------------------------------------------

arg::parse "$@"

# - hexa-admin
#   Runs on localhost:8884
# - hexa-orchestrator
#   Runs on localhost:8885
# - hexa-demo
#   Runs on localhost:8886
# - OPA server
#   Runs on localhost:8887
# - hexa-demo-config
#   Runs on localhost:8889

_main_() {
  pushd ${REPO} > /dev/null
    if [ "${hexa}" == "false" ] && [ "${demo}" == "false" ] ; then
      serve::all
    fi

    if [ "${demo}" == "true" ] ; then
      serve::demo
    fi

    if [ "${hexa}" == "true" ] ; then
      serve::hexa
    fi
  popd > /dev/null
}

serve::all() {
  overmind::start
}

serve::demo() {
  overmind::start --processes demo_cfg,demo_opa,demo_web
}

serve::hexa() {
  overmind::start --processes hexa_adm,hexa_orc
}

overmind::start() {
  overmind start --procfile "${REPO}/bin/dev.d/Procfile" "$@"
}

_main_ "$@"
