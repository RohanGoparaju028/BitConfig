#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-v0.1.0}"
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9.-]+)?$ ]]; then
  echo "Expected a version such as v0.1.0 or v0.1.0-rc.1" >&2
  exit 1
fi
ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
DIST_DIR="$ROOT_DIR/dist"
STAGE_DIR="$(mktemp -d)"
trap 'rm -rf "$STAGE_DIR"' EXIT

mkdir -p "$DIST_DIR"

build_release() {
  local goos="$1"
  local goarch="$2"
  local package="bitconfig-${VERSION}-${goos}-${goarch}"
  local stage="$STAGE_DIR/$package"
  local binary="bitconfig"

  if [[ "$goos" == "windows" ]]; then
    binary="bitconfig.exe"
  fi

  mkdir -p "$stage/gnn"
  echo "-> Building $goos/$goarch..."
  (cd "$ROOT_DIR" && CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -mod=readonly -trimpath -buildvcs=false -ldflags="-s -w" -o "$stage/$binary" ./main.go)
  cp "$ROOT_DIR"/gnn/*.py "$stage/gnn/"
  cp "$ROOT_DIR/gnn/requirements.txt" "$stage/gnn/"
  cp "$ROOT_DIR/README.md" "$ROOT_DIR/LICENSE" "$stage/"
  python3 "$ROOT_DIR/scripts/package_release.py" "$stage" "$DIST_DIR/$package.tar.gz"
}

echo "=== Building BitConfig $VERSION release archives ==="
build_release darwin arm64
build_release darwin amd64
build_release linux amd64
build_release linux arm64
build_release windows amd64

(
  cd "$DIST_DIR"
  shasum -a 256 bitconfig-"$VERSION"-*.tar.gz > SHA256SUMS.txt
)

echo "=== Release archives created in dist/ ==="
ls -lh "$DIST_DIR"
