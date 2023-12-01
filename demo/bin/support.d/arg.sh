#!/usr/bin/env bash

_main_() {
  if [ $# -ge 1 ] ; then
    flag=$1
    shift

    case "${flag}" in
      "--parse" ) _do_parse_ "$@"
                  ;;
      "--usage" ) _do_usage_ "$@"
                  ;;
    esac
  fi
}

# NOTE:
# This is handling a special case in which `--help` was passed as a flag to a
# subcommand, triggering the `docpars` usage, which cannot be eval'd.
_do_parse_() {
  set +e
  if ! eval "$(_parse_args_ "$0" "$@")" 2> /dev/null ; then
    _do_usage_ "$@"
  fi
  set -e
}

_do_usage_() {
  echo "$(_parse_help_ "$0" "$@")"
}

_parse_args_() {
  # NOTE: need to check for `docpars` to cover first-run bootstrapping.
  if command -v docpars > /dev/null ; then
    local script=$1
    shift

    HELP="$(_parse_help_ "$script")"
    docpars -h "$HELP" : "$@"
  fi
}

_parse_help_() {
  local script=$1

  awk '
    BEGIN { u=0; l=0 }
    # we catch the first Usage: match
    /^# Usage:/ {
        if(u == 0)
        {
            u=1
        }
    }

    # match all lines. (Usage: is also matched)
    {
        if(u == 1) {
            # append to an array
            usage[l]=$0
            l++
        }
    }

    # empty line
    /^$/ {
        if(u == 1)
        {
            # stop parsing when empty line found
            u=2
        }
    }

    # display result and format output
    END {
        for(i=0; i<l; i++) {
            # remove comment (see issue #47)
            sub("^# ", "", usage[i])
            sub("^#", "", usage[i])
            sub("^-{48,}", "", usage[i])
            print usage[i]
        }
    }
  ' < "$script"
}

_main_ "$@"
