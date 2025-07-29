#!/bin/bash
# Usage: ./release.sh [sha [tag]]
# Populates release/release.yml with sha/tag (inferring them from HEAD if not specified) and builds binaries
# for several OSs/archs.

SHA=${1:-$(git rev-parse HEAD)}
TAG=${2:-$(git describe --exact-match HEAD 2>/dev/null || echo "")}
SHADATE=$(TZ=UTC0 git show -s --date='format-local:%Y-%m-%dT%H:%M:%SZ' --format="%cd" $SHA)

mkdir -p dist
cat >release/release.yml <<EOF
tag: $TAG
commit: $SHA
date: $SHADATE
EOF

GOOS=linux GOARCH=amd64 go build -o dist/hs.linux
GOOS=darwin GOARCH=amd64 go build -o dist/hs.mac
GOOS=darwin GOARCH=arm64 go build -o dist/hs.m1mac
