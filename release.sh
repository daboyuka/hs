#!/bin/bash
# Usage: ./release.sh [sha [tag]]
# Builds binaries for several OSs/archs, with release info baked in via ldflags.

SHA=${1:-$(git rev-parse HEAD)}
TAG=${2:-$(git describe --exact-match HEAD 2>/dev/null || echo "")}
SHADATE=$(TZ=UTC0 git show -s --date='format-local:%Y-%m-%dT%H:%M:%SZ' --format="%cd" $SHA)


RPKG=github.com/daboyuka/hs/release
LDFLAGS=(
  -ldflags "-X $RPKG.Tag=$TAG -X $RPKG.Commit=$SHA -X $RPKG.DateRaw=$SHADATE"
)

GOOS=linux GOARCH=amd64 go build "${LDFLAGS[@]}" -o dist/hs.linux
GOOS=darwin GOARCH=amd64 go build "${LDFLAGS[@]}" -o dist/hs.mac
GOOS=darwin GOARCH=arm64 go build "${LDFLAGS[@]}" -o dist/hs.m1mac
