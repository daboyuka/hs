#!/bin/bash
# Usage: ./release.sh [sha [tag]]
# Builds binaries for several OSs/archs, with release info baked in via ldflags.

function mkrelease() {
  GOOS=$1 GOARCH=$2 go build -o dist/hs_${1}_${2}
}

mkrelease linux amd64
mkrelease darwin amd64
mkrelease darwin arm64
