#!/usr/bin/env bash
set -euo pipefail

if [ "${1:-}" = "" ]; then
  echo "Usage: scripts/bump_version.sh <version>"
  exit 1
fi

version="$1"

files=(
  "configs/config.yaml"
  "README.md"
  "README.zh-TW.md"
)

for file in "${files[@]}"; do
  if [ ! -f "$file" ]; then
    echo "File not found: $file"
    exit 1
  fi
  perl -pi -e "s/^version:\\s*\".*?\"(.*)\$/version: \"${version}\"\\1/" "$file"
done

echo "Updated version to ${version} in:"
printf "  - %s\n" "${files[@]}"
