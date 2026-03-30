#!/usr/bin/env bash
set -euo pipefail

CORPUS_DIR="${1:-fuzz-corpus}"
MAX_SIZE="${2:-524288}" # 512KB

mkdir -p "$CORPUS_DIR"

for f in parser/testdata/*.jfr.gz; do
    size=$(stat -c%s "$f" 2>/dev/null || stat -f%z "$f" 2>/dev/null)
    if [ "$size" -le "$MAX_SIZE" ]; then
        name=$(basename "$f" .jfr.gz)
        gunzip -c "$f" > "$CORPUS_DIR/$name.jfr"
    fi
done
