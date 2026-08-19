#!/usr/bin/env bash

# Sourced by scripts/stop-checks.sh. Reads/writes the orchestrator's globals
# ($go_changed, $out, $fail) directly — that's the point of sourcing, not exporting.

run_go_checks() {
  [ -z "$go_changed" ] && return

  if ! go_out=$(go build ./... 2>&1); then
    out+="go build failed:\n$go_out\n\n"
    fail=1
  fi

  if ! gotest_out=$(go test ./... 2>&1); then
    out+="go test failed:\n$gotest_out\n\n"
    fail=1
  fi
  # go vet + staticcheck. staticcheck COMPILES the whole module, so it is

  if ! sc_out=$(bash scripts/checks/lang/check-staticcheck.sh 2>&1); then
    out+="check-staticcheck failed:\n$sc_out\n\n"
    fail=1
  fi
}
