#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE_MEDIA_DIR="${ROOT_DIR}/source_media"
SOURCE_MEDIA_SORT_DIR="${ROOT_DIR}/source_media_sort"
SOURCE_MEDIA_COPY_DIR="${ROOT_DIR}/source_media_copy"

# Remove existing test data directories.
rm -rf "${SOURCE_MEDIA_DIR}" "${SOURCE_MEDIA_SORT_DIR}"

# Recreate source_media by copying from source_media_copy.
if [[ -d "${SOURCE_MEDIA_COPY_DIR}" ]]; then
  cp -a "${SOURCE_MEDIA_COPY_DIR}" "${SOURCE_MEDIA_DIR}"
else
  echo "source_media_copy not found: ${SOURCE_MEDIA_COPY_DIR}" >&2
  exit 1
fi
