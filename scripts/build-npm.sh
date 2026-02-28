#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

platforms=(
  "darwin  arm64 nsprobe-darwin-arm64 nsprobe"
  "darwin  amd64 nsprobe-darwin-x64   nsprobe"
  "linux   amd64 nsprobe-linux-x64    nsprobe"
  "linux   arm64 nsprobe-linux-arm64  nsprobe"
  "windows amd64 nsprobe-win32-x64    nsprobe.exe"
)

for entry in "${platforms[@]}"; do
  read -r goos goarch pkg bin <<< "$entry"
  outdir="$ROOT/npm/$pkg/bin"
  mkdir -p "$outdir"
  echo "Building $pkg (GOOS=$goos GOARCH=$goarch)..."
  GOOS="$goos" GOARCH="$goarch" go build -o "$outdir/$bin" "$ROOT"
done

echo "Done. All binaries built."
