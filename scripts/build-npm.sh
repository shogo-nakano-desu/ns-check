#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

platforms=(
  "darwin  arm64 ns-check-darwin-arm64 ns-check"
  "darwin  amd64 ns-check-darwin-x64   ns-check"
  "linux   amd64 ns-check-linux-x64    ns-check"
  "linux   arm64 ns-check-linux-arm64  ns-check"
  "windows amd64 ns-check-win32-x64    ns-check.exe"
)

for entry in "${platforms[@]}"; do
  read -r goos goarch pkg bin <<< "$entry"
  outdir="$ROOT/npm/$pkg/bin"
  mkdir -p "$outdir"
  echo "Building $pkg (GOOS=$goos GOARCH=$goarch)..."
  GOOS="$goos" GOARCH="$goarch" go build -o "$outdir/$bin" "$ROOT"
done

echo "Done. All binaries built."
