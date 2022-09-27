#!/usr/bin/env bash
set -euo pipefail
source ${REPO}/bin/support.sh

# ------------------------------------------------------------------------------
# Usage:
#
#   serve [options]
#
# Description:
#
#   Deploy and run the Hexa applications (in Docker).
#
# Options:
#
#   -h, --help                Print help text.
# ------------------------------------------------------------------------------

arg::parse "$@"

_main_() {
  pushd ${REPO} > /dev/null
    exec::step "Prepare development database" "
      chmod 775 ${REPO}/databases/docker_support/initdb.d/create-databases.sh
      chmod 775 ${REPO}/databases/docker_support/migrate-databases.sh
      chmod 600 ${REPO}/databases/docker_support/ca-cert.pem
      chmod 600 ${REPO}/databases/docker_support/client-cert.pem
      chmod 600 ${REPO}/databases/docker_support/client-key.pem
    "
    exec::step "Deploy Hexa applications in 'Docker'" "docker-compose up"
  popd > /dev/null
}

_main_ "$@"
