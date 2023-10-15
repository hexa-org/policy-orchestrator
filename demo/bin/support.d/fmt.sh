#!/usr/bin/env bash

fmt::indent() {
  set +u
  local indent=1
  if [ -n "$1" ]; then
    indent=$1
  fi
  pr -to $(($indent))
  set -u
}

fmt::grid() {
  bat --paging=never --style grid,numbers
}
