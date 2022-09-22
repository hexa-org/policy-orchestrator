#!/usr/bin/env bash
set -euo pipefail
source ${REPO}/bin/support.sh

# ------------------------------------------------------------------------------
# Usage:
#
#   install [options]
#
# Description:
#
#   Install workspace/dev dependencies.
#
# Options:
#
#   -t, --target=<name>       Installation package or manager [default: all].
#                             One of:
#                               - all (default)
#                               - asdf
#                               - opa
#   -h, --help                Print help text.
# ------------------------------------------------------------------------------

arg::parse "$@"

_main_() {
  case "${target}" in
    all )
      ensure::all
      ;;
    asdf )
      ensure::asdf
      ;;
    opa )
      ensure::opa
      ;;
    * )
      echo::fail "unknown install target: '${target}'"
      ;;
  esac
}

ensure::all() {
  ensure::asdf
  ensure::opa
}

# asdf
# ----------------------------------------------------------
ensure::asdf() {
  exec::stat "asdf" "command -v asdf" "install::asdf"
}

install::asdf() {
  brew install asdf
  # OR: "open 'https://asdf-vm.com/#/core-manage-asdf'"
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
