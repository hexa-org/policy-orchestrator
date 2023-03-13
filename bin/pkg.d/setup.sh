#!/usr/bin/env bash
set -euo pipefail

REPO=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )/../.." &> /dev/null && pwd )
source ${REPO}/bin/support.sh

# ------------------------------------------------------------------------------
# Usage:
#
#   setup [options]
#
# Description:
#
#   Install/configure runtime dependencies.
#
# Options:
#
#   -h, --help                Print help text.
# ------------------------------------------------------------------------------

arg::parse "$@"

_main_() {
  pushd ${REPO} > /dev/null
    ensure::brew
    ensure::asdf
    ensure::path
  popd > /dev/null
}

# homebrew & packages
# ----------------------------------------------------------
ensure::brew() {
  command -v brew > /dev/null || setup::brew
  brew bundle
}

setup::brew() {
  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
}

# asdf packages (just golang)
# ----------------------------------------------------------
ensure::asdf() {
  (go version | grep 'go1.19' > /dev/null) || setup::asdf
}

setup::asdf() {
  asdf install golang
}

# path (i.e., direnv)
# ----------------------------------------------------------
ensure::path() {
  direnv allow
}

_main_ "$@"
