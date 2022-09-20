#!/usr/bin/env bash

echo::info() {
  echo::color --green "[info]" --reset "$@"
}

echo::warn() {
  echo::color --yellow "[warn]" --reset "$@"
}

echo::fail() {
  echo::color --red "[fail]" --reset "$@"
  exit 1
}

echo::step() {
  echo::color --green "     •" --reset "$@"
}

echo::color() {
  string=""

  if [[ $1 == "-n" ]] ; then
    line_ending=""
  else
    line_ending="\n"
  fi

  while (( "$#" )); do
    case "$1" in
      -n)
        shift
        ;;

      --reset)
        string="${string}\033[0m"
        shift
        ;;

      --*)
        color="\033["

        if [[ $1 == *"-bold"* ]]; then
          color="${color}1;"
        else
          color="${color}0;"
        fi

        if [[ $1 == *"-bright"* ]]; then
          color="${color}9"
        else
          color="${color}3"
        fi

        case "$1" in
          *-blue*)
            color="${color}4"
            ;;

          *-cyan*)
            color="${color}6"
            ;;

          *-green*)
            color="${color}2"
            ;;

          *-magenta*)
            color="${color}5"
            ;;

          *-red*)
            color="${color}1"
            ;;

          *-white*)
            color="${color}7"
            ;;

          *-yellow*)
            color="${color}3"
            ;;
        esac

        color="${color}m"

        string="${string}${color}"
        shift
        ;;

      *)
        string="${string}${1} "
        shift
        ;;
    esac
  done

  echo -n -e "${string}\033[0m${line_ending}"
}
