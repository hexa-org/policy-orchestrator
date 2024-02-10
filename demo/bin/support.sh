#!/usr/bin/env bash

source ${REPO}/bin/support.d/echo.sh
source ${REPO}/bin/support.d/exec.sh
source ${REPO}/bin/support.d/fmt.sh

arg::parse() {
  source ${REPO}/bin/support.d/arg.sh --parse "$@"
}

arg::usage() {
  source ${REPO}/bin/support.d/arg.sh --usage "$@"
}

cmd::proxy() {
  source ${REPO}/bin/support.d/cmd.sh --proxy "$@"
}
