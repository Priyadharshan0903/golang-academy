#!/usr/bin/env bash
# For "test your tests" exercises: every bug_*.go variant must make the
# reference tests (build tag "solution") fail. Run from the repo root.
set -euo pipefail
cd "$(dirname "$0")/../exercises"

status=0
for dir in $(find . -name 'bug_*.go' -exec dirname {} \; | sort -u); do
  if ! go test -count=1 -tags solution "$dir" > /dev/null; then
    echo "error: reference tests fail on the correct code in $dir" >&2
    status=1
  fi
  for f in "$dir"/bug_*.go; do
    tag=$(basename "$f" .go)
    if go test -count=1 -tags "solution $tag" "$dir" > /dev/null 2>&1; then
      echo "error: reference tests in $dir do not catch $tag" >&2
      status=1
    else
      echo "ok   $dir catches $tag"
    fi
  done
done
exit $status
