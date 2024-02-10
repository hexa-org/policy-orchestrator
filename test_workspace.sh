#!/bin/bash


prefix="github.com/hexa-org/policy-orchestrator"
go mod graph | awk '{print $1}' | cut -d '@' -f 1 | sort | uniq | grep "policy-orchestrator" |  while read x; do
  echo $x
  suffix_removed=${x/#$prefix}
  use_mod=".$suffix_removed"
  echo $use_mod
  go test -tags integration -coverprofile coverage.out "${use_mod}/.../"
  #go test "${use_mod}/.../"
done
