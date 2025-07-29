#!/bin/bash
# Usage: ./release.sh [sha [tag]]
# Builds binaries for several OSs/archs, with release info baked in via ldflags.

SHA=${1:-$(git rev-parse HEAD)}
TAG=${2:-$(git describe --exact-match --tags HEAD 2>/dev/null || echo "")}
SHADATE=$(TZ=UTC0 git show -s --date='format-local:%Y-%m-%dT%H:%M:%SZ' --format="%cd" $SHA)

RPKG=github.com/daboyuka/hs/release
LDFLAGS=(
  -ldflags "-X $RPKG.Tag=$TAG -X $RPKG.Commit=$SHA -X $RPKG.DateRaw=$SHADATE"
)

function mkrelease() {
  GOOS=$1 GOARCH=$2 go build "${LDFLAGS[@]}" -o dist/hs_${1}_${2}
}

mkrelease linux amd64
mkrelease darwin amd64
mkrelease darwin arm64
