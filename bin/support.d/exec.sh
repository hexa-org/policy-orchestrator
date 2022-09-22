#!/usr/bin/env bash

exec::fail() {
  echo::fail "\n\nUsage:$@"
}

exec::stat() {
  label=$1
  check=$2
  build=$3
  trace="${REPO}/.local/setup.out"

  echo::step -n "check ${label}:"

  flags=$-
  set +e

  eval "${check} > ${trace} 2>&1"
  status=$?

  if [[ ${flags} =~ e ]] ; then
    set -e
  else
    set +e
  fi

  if [ ${status} -eq 0 ]; then
    echo::color --green "OK"
  else
    echo::color --red "FAILED"
    # cat ${trace}
    echo::info "build ${label}..."
    eval "${build}"
  fi

  rm ${trace}
  return ${status}
}

exec::step() {
  usage="
    exec::step \"<command>\"
    exec::step \"<message>\" \"<command>\"
  "

  _main_() {
    _check_ "$#"

    local cmd="$1"
    local msg="$1"

    if [ "$#" == "2" ] ; then
      cmd="$2"
    fi

    echo::step "${msg}"
    eval "${cmd}" 2>&1 | fmt::grid
  }

  _check_() {
    if [ "$1" -lt "1" ] || [ "$1" -gt "2" ] ; then
      exec::fail "${usage}"
    fi
  }

  _main_ "$@"
}
