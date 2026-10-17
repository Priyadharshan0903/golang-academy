#!/usr/bin/env bash
# House style checks that are not covered by gofmt or go vet.
set -euo pipefail

cd "$(dirname "$0")/.."

status=0

# 1. No em dashes anywhere in our own files (house style, see plan/decisions.md D9).
em_dash=$(printf '\xe2\x80\x94')
if grep -rIl --exclude-dir={.git,public,resources,node_modules,lib} "$em_dash" . ; then
  echo "error: em dash found in the files above; use a comma, colon or parentheses instead" >&2
  status=1
fi

# 2. Every Go file is gofmt-formatted.
unformatted=$(gofmt -l examples exercises projects deploy)
if [ -n "$unformatted" ]; then
  echo "error: files not formatted with gofmt:" >&2
  echo "$unformatted" >&2
  status=1
fi

exit $status
