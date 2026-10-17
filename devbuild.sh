#!/usr/bin/env bash
# ScraperGo: build standalone Intel and Apple Silicon macOS executables.
set -Eeuo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

if ! command -v go >/dev/null 2>&1; then
  echo 'Error: Go must be installed on the build machine.' >&2
  exit 1
fi
if [[ ! -f go.mod ]]; then
  echo "Error: go.mod not found in $SCRIPT_DIR. Place devbuild.sh in the ScraperGo project root." >&2
  exit 1
fi

OUT_DIR="$SCRIPT_DIR/dist"
mkdir -p "$OUT_DIR"

printf 'ScraperGo build with %s\n' "$(go version)"
printf 'Project: %s\n' "$SCRIPT_DIR"

# Fail fast if local source does not pass the test suite.
printf '\nRunning tests...\n'
go test ./...

build_target() {
  local arch="$1"
  local output="$OUT_DIR/scrapergo-darwin-$arch"
  printf '\nBuilding darwin/%s...\n' "$arch"
  CGO_ENABLED=0 GOOS=darwin GOARCH="$arch" \
    go build -trimpath -buildvcs=false -ldflags='-s -w' -o "$output" .
  chmod 755 "$output"
  printf 'Created: %s\n' "$output"
  if command -v file >/dev/null 2>&1; then
    file "$output"
  fi
}

build_target amd64
build_target arm64

printf '\nBuild complete. Files in %s:\n' "$OUT_DIR"
ls -lh "$OUT_DIR/scrapergo-darwin-amd64" "$OUT_DIR/scrapergo-darwin-arm64"
printf '\nOn an Intel Mac: ./dist/scrapergo-darwin-amd64\n'
printf 'On an Apple Silicon Mac: ./dist/scrapergo-darwin-arm64\n'
printf 'Pass --port 8081 to either executable to select another port.\n'
