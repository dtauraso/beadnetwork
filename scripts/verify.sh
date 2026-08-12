#!/usr/bin/env bash















set -u
exec bash "$(dirname "$0")/stop-checks.sh" --cli "$@"
