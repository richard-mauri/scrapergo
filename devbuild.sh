#!/usr/bin/env bash
# ScraperGo: build standalone macOS Intel/Apple Silicon binaries; optionally publish a GitHub Release.
set -Eeuo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"
REPO="richard-mauri/scrapergo"
OUT_DIR="$SCRIPT_DIR/dist"
VERSION=""

usage() {
  cat <<'HELP'
Usage:
  ./devbuild.sh                 Build both macOS binaries and SHA256SUMS
  ./devbuild.sh --release vX.Y.Z  Build and publish a new GitHub Release
  ./devbuild.sh --help          Display this help

Requires Go for building. Publishing additionally requires an authenticated
GitHub CLI (gh), a clean Git working tree, and the current HEAD pushed to its
upstream branch. A release with the requested tag must not already exist.

Release assets: dist/scrapergo-darwin-amd64,
                dist/scrapergo-darwin-arm64, dist/SHA256SUMS
HELP
}

case "${1:-}" in
  "") [[ $# -eq 0 ]] || { usage >&2; exit 2; } ;;
  -h|--help) [[ $# -eq 1 ]] || { usage >&2; exit 2; }; usage; exit 0 ;;
  --release)
    [[ $# -eq 2 ]] || { echo 'Error: expected --release vX.Y.Z' >&2; usage >&2; exit 2; }
    VERSION="$2"
    if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
      echo 'Error: release version must look like v0.2.3' >&2; exit 2
    fi ;;
  *) echo "Error: unknown option: $1" >&2; usage >&2; exit 2 ;;
esac

command -v go >/dev/null 2>&1 || { echo 'Error: Go is not installed.' >&2; exit 1; }
[[ -f go.mod ]] || { echo "Error: missing go.mod in $SCRIPT_DIR" >&2; exit 1; }

if [[ -n "$VERSION" ]]; then
  for cmd in gh git; do
    command -v "$cmd" >/dev/null 2>&1 || { echo "Error: $cmd is required to publish a release." >&2; exit 1; }
  done
  gh auth status >/dev/null 2>&1 || { echo 'Error: authenticate with gh auth login first.' >&2; exit 1; }
  [[ -z "$(git status --porcelain)" ]] || { echo 'Error: commit or stash changes before publishing a release.' >&2; exit 1; }
  branch="$(git symbolic-ref --quiet --short HEAD || true)"
  [[ "$branch" == main ]] || { echo 'Error: publish releases from the main branch.' >&2; exit 1; }
  upstream="$(git rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' 2>/dev/null || true)"
  [[ -n "$upstream" ]] || { echo 'Error: no upstream configured. Run git push -u origin main.' >&2; exit 1; }
  # A local tracking ref might be stale; require its SHA to match, and remind user to push.
  [[ "$(git rev-parse HEAD)" == "$(git rev-parse '@{upstream}')" ]] || {
    echo 'Error: local HEAD differs from its upstream. Commit and git push first.' >&2; exit 1;
  }
  if gh release view "$VERSION" --repo "$REPO" >/dev/null 2>&1; then
    echo "Error: GitHub Release $VERSION already exists. Choose a new version." >&2; exit 1
  fi
  if git ls-remote --exit-code --tags origin "refs/tags/$VERSION" >/dev/null 2>&1; then
    echo "Error: Git tag $VERSION already exists on origin. Choose a new version." >&2; exit 1
  fi
fi

mkdir -p "$OUT_DIR"
printf 'ScraperGo build: %s\n' "$(go version)"
printf 'Running tests...\n'
go test ./...

for arch in amd64 arm64; do
  output="$OUT_DIR/scrapergo-darwin-$arch"
  printf 'Building darwin/%s...\n' "$arch"
  CGO_ENABLED=0 GOOS=darwin GOARCH="$arch" \
    go build -trimpath -buildvcs=false -ldflags='-s -w' -o "$output" .
  chmod 755 "$output"
  command -v file >/dev/null 2>&1 && file "$output"
done

(
  cd "$OUT_DIR"
  shasum -a 256 scrapergo-darwin-amd64 scrapergo-darwin-arm64 > SHA256SUMS
)
printf '\nBuilt binaries and checksums in %s\n' "$OUT_DIR"
ls -lh "$OUT_DIR/scrapergo-darwin-amd64" "$OUT_DIR/scrapergo-darwin-arm64" "$OUT_DIR/SHA256SUMS"

if [[ -n "$VERSION" ]]; then
  printf '\nPublishing GitHub Release %s...\n' "$VERSION"
  gh release create "$VERSION" \
    "$OUT_DIR/scrapergo-darwin-amd64" \
    "$OUT_DIR/scrapergo-darwin-arm64" \
    "$OUT_DIR/SHA256SUMS" \
    --repo "$REPO" --target main \
    --title "ScraperGo $VERSION" \
    --generate-notes
  printf '\nRelease published: https://github.com/%s/releases/tag/%s\n' "$REPO" "$VERSION"
else
  printf '\nTo publish after committing and pushing changes:\n  ./devbuild.sh --release vX.Y.Z\n'
fi
