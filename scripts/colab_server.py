#!/usr/bin/env python3
"""
Google Colab GPU LLM Server Provisioner for Novel-TUI.

This script provisions a Google Colab T4 GPU instance, downloads KoboldCpp with CUDA acceleration,
fetches the Llama-3-8B-Stheno-v3.2-Q5_K_M GGUF model from Hugging Face, launches the server
with a Cloudflare tunnel, and prints the generated OpenAI-compatible endpoint URL.
"""

import os
import re
import shutil
import subprocess
import sys
import time

SESSION_NAME = "novel-llm"
GPU_TYPE = "T4"
KOBOLD_URL = "https://github.com/LostRuins/koboldcpp/releases/download/v1.73/koboldcpp-linux-x64-cuda1200"
MODEL_URL = "https://huggingface.co/bartowski/Llama-3-8B-Stheno-v3.2-GGUF/resolve/main/Llama-3-8B-Stheno-v3.2-Q5_K_M.gguf"

CLOUDFLARE_REGEX = re.compile(r"https://[a-zA-Z0-9-]+\.trycloudflare\.com(?:/v1)?")


def check_colab_cli() -> bool:
    """Checks if the colab CLI is installed and available in PATH."""
    return shutil.which("colab") is not None


def run_command(cmd: list[str], check: bool = True) -> subprocess.CompletedProcess:
    """Runs a system command and returns the completed process."""
    return subprocess.run(cmd, capture_output=True, text=True, check=check)


def create_session() -> bool:
    """Creates a new Colab session with GPU T4 if not already running."""
    print(f"[Colab] Provisioning session '{SESSION_NAME}' with GPU {GPU_TYPE}...")
    try:
        # Check if session exists or create a new one
        res = subprocess.run(
            ["colab", "new", "--gpu", GPU_TYPE, "--session", SESSION_NAME],
            capture_output=True,
            text=True,
        )
        if res.returncode != 0 and "already exists" not in res.stderr.lower():
            print(f"[Colab] Warning creating session: {res.stderr.strip()}", file=sys.stderr)
        return True
    except Exception as e:
        print(f"[Colab] Error launching session: {e}", file=sys.stderr)
        return False


def start_remote_server():
    """Runs the setup and launches KoboldCpp on the remote Colab instance."""
    setup_commands = f"""
set -e
echo '[Colab] Preparing workspace...'
mkdir -p /content/novel-llm
cd /content/novel-llm

if [ ! -f koboldcpp ]; then
    echo '[Colab] Downloading KoboldCpp CUDA binary...'
    curl -fLo koboldcpp '{KOBOLD_URL}'
    chmod +x koboldcpp
fi

if [ ! -f model.gguf ]; then
    echo '[Colab] Downloading Llama-3-8B-Stheno-v3.2 model...'
    curl -fLo model.gguf '{MODEL_URL}'
fi

echo '[Colab] Launching KoboldCpp with GPU layers and Cloudflare tunnel...'
./koboldcpp --model model.gguf --usecublas --gpulayers 33 --contextsize 8192 --usecloudflare --skiplauncher
"""

    print("[Colab] Executing setup on Colab instance...")
    process = subprocess.Popen(
        ["colab", "run", "--session", SESSION_NAME, "--", "bash", "-c", setup_commands],
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        bufsize=1,
    )

    tunnel_url = None
    if process.stdout is not None:
        for line in process.stdout:
            sys.stdout.write(line)
            sys.stdout.flush()

            match = CLOUDFLARE_REGEX.search(line)
            if match:
                url = match.group(0)
                if not url.endswith("/v1"):
                    url = url.rstrip("/") + "/v1"
                tunnel_url = url
                print(f"\n[Colab] ✅ Successfully established Cloudflare endpoint: {tunnel_url}\n")
                sys.stdout.flush()
                # Server is live, we can keep running in background or monitor
                break

    if not tunnel_url:
        process.wait()
        if process.returncode != 0:
            print(f"[Colab] Process exited with error code {process.returncode}", file=sys.stderr)
            sys.exit(process.returncode)


def main():
    if len(sys.argv) > 1 and sys.argv[1] in ("--check", "check"):
        if check_colab_cli():
            print("colab CLI is installed.")
            sys.exit(0)
        else:
            print("colab CLI is NOT installed.", file=sys.stderr)
            sys.exit(1)

    if len(sys.argv) > 1 and sys.argv[1] in ("--stop", "stop"):
        print(f"[Colab] Stopping session '{SESSION_NAME}'...")
        subprocess.run(["colab", "stop", "--session", SESSION_NAME])
        sys.exit(0)

    if not check_colab_cli():
        print(
            "Error: 'colab' CLI is not found in PATH.\n"
            "Please install it using: pip install google-colab-cli\n"
            "And authenticate with: colab auth login",
            file=sys.stderr,
        )
        sys.exit(1)

    create_session()
    start_remote_server()


if __name__ == "__main__":
    main()
