#!/usr/bin/env bash
# The check before a change is done (CLAUDE.md, working style): formatting, vet,
# and the whole suite uncached with the race detector, against the real
# database — a missing database is a failure here, never a skip. Runs on
# every push through .githooks/pre-push; no hosted CI, no credits.
set -euo pipefail
cd "$(dirname "$0")/.."
if [ "$(git config core.hooksPath || true)" != ".githooks" ]; then
  echo "note: the pre-push hook is not wired in this clone — run: git config core.hooksPath .githooks" >&2
fi
unformatted=$(gofmt -l cmd internal markets)
if [ -n "$unformatted" ]; then
  echo "gofmt: $unformatted" >&2
  exit 1
fi
go vet ./...
RHEA_CI=1 go test ./... -count=1 -race -timeout 20m
