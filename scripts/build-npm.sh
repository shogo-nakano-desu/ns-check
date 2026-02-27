#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

platforms=(
  "darwin  arm64 nmchk-darwin-arm64 nmchk"
  "darwin  amd64 nmchk-darwin-x64   nmchk"
  "linux   amd64 nmchk-linux-x64    nmchk"
  "linux   arm64 nmchk-linux-arm64  nmchk"
  "windows amd64 nmchk-win32-x64    nmchk.exe"
)

for entry in "${platforms[@]}"; do
  read -r goos goarch pkg bin <<< "$entry"
  outdir="$ROOT/npm/$pkg/bin"
  mkdir -p "$outdir"
  echo "Building $pkg (GOOS=$goos GOARCH=$goarch)..."
  GOOS="$goos" GOARCH="$goarch" go build -o "$outdir/$bin" "$ROOT"
done

echo "Done. All binaries built."
