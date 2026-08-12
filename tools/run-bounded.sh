#!/usr/bin/env bash

set -uo pipefail

if [ "$#" -lt 2 ]; then
  echo "usage: run-bounded.sh <seconds> <command> [args...]" >&2
  exit 2
fi

limit="$1"; shift

perl -e '
  my $limit = shift;
  my $pid = fork();
  if ($pid == 0) { exec @ARGV or exit 127; }
  $SIG{ALRM} = sub { kill "TERM", $pid; sleep 1; kill "KILL", $pid; exit 124; };
  alarm $limit;
  waitpid($pid, 0);
  exit($? >> 8);
' "$limit" "$@"
