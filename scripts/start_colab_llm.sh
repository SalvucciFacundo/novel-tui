#!/usr/bin/env bash
# Wrapper script to start the Google Colab GPU LLM server for Novel-TUI.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_BIN="python3"

if ! command -v "$PYTHON_BIN" >/dev/null 2>&1; then
    echo "Error: python3 is required to run colab_server.py" >&2
    exit 1
fi

exec "$PYTHON_BIN" "${SCRIPT_DIR}/colab_server.py" "$@"
