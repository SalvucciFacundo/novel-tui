#!/usr/bin/env bash
# Regenerates the embedded spellcheck wordlists.
# Requires: curl, sort, gzip, unmunch (hunspell tools).
set -euo pipefail
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
OUT="internal/service/spell/data"

echo "-> English (dwyl/english-words)"
curl -sL -o "$TMP/en.txt" https://raw.githubusercontent.com/dwyl/english-words/master/words_alpha.txt
tr -d '\r' < "$TMP/en.txt" | sort -u | gzip -9 > "$OUT/en.txt.gz"

echo "-> Spanish (titoBouzout/Dictionaries, expanded via unmunch)"
curl -sL -o "$TMP/Spanish.dic" https://raw.githubusercontent.com/titoBouzout/Dictionaries/master/Spanish.dic
curl -sL -o "$TMP/Spanish.aff" https://raw.githubusercontent.com/titoBouzout/Dictionaries/master/Spanish.aff
unmunch "$TMP/Spanish.dic" "$TMP/Spanish.aff" 2>/dev/null | sort -u | gzip -9 > "$OUT/es.txt.gz"

echo "done:"; ls -la "$OUT"/*.gz
