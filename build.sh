#!/usr/bin/env bash

set -euo pipefail

case "${1:-build}" in
  build)
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w"
    ;;
  clean)
    rm -f "$(basename "$PWD")"
    ;;
  *)
    echo "usage: $0 [build|clean]" >&2
    exit 1
    ;;
esac
