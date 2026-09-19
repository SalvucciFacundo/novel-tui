# Dictionary sources (embedded as .txt.gz, loaded at runtime)

- English (`en.txt.gz`): `words_alpha.txt` from dwyl/english-words
  (https://github.com/dwyl/english-words), CRLF stripped, sorted, deduped.
- Spanish (`es.txt.gz`): `Spanish.dic` + `Spanish.aff` from titoBouzout/Dictionaries
  (LibreOffice hunspell dictionaries), expanded with `unmunch`, sorted, deduped.

Regenerate with `scripts/fetch_dictionaries.sh` (requires curl, sort, gzip, unmunch).
