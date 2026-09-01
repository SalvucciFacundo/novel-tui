# 📖 Novel-TUI

[![Go Version](https://img.shields.io/github/go-mod/go-version/SalvucciFacundo/novel-tui)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/SalvucciFacundo/novel-tui?color=brightgreen)](https://github.com/SalvucciFacundo/novel-tui/releases)

> **Novel-TUI** is a distraction-free, terminal-based text editor and creative writing suite specifically designed for light novel and fiction authors. Built in **Go** with **Bubble Tea** (The Elm Architecture) and **Lip Gloss**.

---

## 📸 Overview

### Launcher / Dashboard

```text
  ███╗   ██╗ ██████╗ ██╗   ██╗███████╗██╗         ████████╗██╗   ██╗██╗
  ████╗  ██║██╔═══██╗██║   ██║██╔════╝██║         ╚══██╔══╝██║   ██║██║
  ██╔██╗ ██║██║   ██║██║   ██║█████╗  ██║   █████╗   ██║   ██║   ██║██║
  ██║╚██╗██║██║   ██║╚██╗ ██╔╝██╔══╝  ██║   ╚════╝   ██║   ██║   ██║██║
  ██║ ╚████║╚██████╔╝ ╚████╔╝ ███████╗███████╗       ██║   ╚██████╔╝██║
  ╚═╝  ╚═══╝ ╚═════╝   ╚═══╝  ╚══════╝╚══════╝       ╚═╝    ╚═════╝ ╚═╝
             — Entorno de Escritura para Novelas y Ficción —

 ┌─ Novelas Recientes ────────────────────┐  ┌─ Menú Principal ───────────────┐
 │  > 📖 El Despertar del Rey (12 cap)    │  │  [c] ⚡ Continuar última novela│
 │    📖 Crónicas del Abismo (5 cap)      │  │  [n] ✨ Nueva novela           │
 │    📖 Sombras en el Valle (1 cap)      │  │  [o] 📂 Abrir otra carpeta     │
 │                                        │  │  [l] 🤖 Configurar LLM / IA    │
 │                                        │  │  [d] 📁 Cambiar Raíz (~/Novelas│
 │                                        │  │  [q] 🚪 Salir                  │
 └────────────────────────────────────────┘  └────────────────────────────────┘
  📁 Raíz: /home/kuno/Novelas  |  🤖 LLM: Ollama (qwen2.5:7b)  |  [Enter: Abrir]
```

### Editor de Novela

```text
┌─ [Chapters] [Lore / Characters] ───────┬─ [Capítulo 01: El Comienzo] ──────────────┐
│ > 01. El Comienzo                      │ El viento soplaba con fuerza sobre las    │
│   02. Sombras en el Valle              │ colinas de piedra. Kuno observó la figura │
│                                        │ que descendía lentamente hacia el sendero │
│                                        │ iluminado por la luna...                  │
│                                        │                                           │
│                                        │                                           │
│ [Ctrl+N / n: New Chapter]              │                                           │
└────────────────────────────────────────┴─ [Capítulo 01] | 340 words | ~2 min read ─┘
```

---

## ✨ Features

- 🧘 **Distraction-Free Writing**: Terminal-native writing environment without heavy GUI bloat or complex configs.
- 📚 **Chapter Management**: Dedicated sidebar for organizing, reordering, and switching chapters instantly.
- 👥 **Character & Worldbuilding Lore**: Integrated character cards (role, backstory, notes) viewable directly alongside the editor without losing your place.
- ⏱️ **Real-Time Writing Metrics**: Live word count, character count, estimated reading time, and clean `[Saved]` / `[Modified*]` status indicators.
- 🛡️ **Atomic File Persistence**: Never lose your work. Chapters and character data are persisted safely on local disk using atomic file transactions.
- 🎨 **Modern Terminal Aesthetics**: Styled with the Catppuccin palette via Lip Gloss, with clear focus indicators, responsive resizing, and small-screen safety guards.

---

## 🚀 Installation

### 1. Via Go (Recommended for Go developers)

Install the latest binary directly to your `$GOPATH/bin`:

```bash
go install github.com/SalvucciFacundo/novel-tui/cmd/novel-tui@latest
```

Ensure `$(go env GOPATH)/bin` is in your `$PATH`, then run:

```bash
novel-tui
```

---

### 2. Linux Packages & Pre-built Binaries

Download pre-compiled binaries for your architecture from the [Releases](https://github.com/SalvucciFacundo/novel-tui/releases) page.

#### One-Line Binary Install (Linux x86_64 / arm64)

```bash
curl -sSL https://raw.githubusercontent.com/SalvucciFacundo/novel-tui/main/scripts/install.sh | bash
```

#### Debian / Ubuntu (.deb)

```bash
# Download .deb from Releases
sudo dpkg -i novel-tui_*_linux_amd64.deb
```

#### Fedora / RHEL (.rpm)

```bash
# Download .rpm from Releases
sudo rpm -i novel-tui-*_linux_amd64.rpm
```

#### Arch Linux (Standalone binary)

```bash
tar -xzf novel-tui_*_linux_amd64.tar.gz
sudo mv novel-tui /usr/local/bin/
```

---

### 3. Build from Source

```bash
git clone https://github.com/SalvucciFacundo/novel-tui.git
cd novel-tui
go build -o novel-tui ./cmd/novel-tui
sudo mv novel-tui /usr/local/bin/
```

---

## ⌨️ Keybindings

| Keybinding | Scope | Description |
|---|---|---|
| `Tab` | Global | Cycle focus forward (Sidebar ↔ Editor) |
| `Shift + Tab` | Global | Cycle focus backward (Editor ↔ Sidebar) |
| `Ctrl + S` | Editor / Global | Save current chapter atomically to disk |
| `Ctrl + N` / `n` | Sidebar (Chapters) | Create a new chapter |
| `[` / `]` or `h` / `l` | Sidebar | Toggle sidebar tabs between **Chapters** and **Lore / Characters** |
| `Up` / `Down` or `k` / `j` | Sidebar | Navigate chapters or character cards |
| `Enter` | Sidebar (Chapters) | Open and load selected chapter into editor |
| `Ctrl + C` / `Esc` | Global | Exit application |

---

## 📁 Workspace File Structure

When you launch `novel-tui` in any directory, it initializes a clean local workspace:

```text
my-novel/
├── chapters/
│   ├── 01_capitulo_1.md
│   ├── 02_capitulo_2.md
│   └── ...
└── characters.json
```

All data is stored in standard Markdown and JSON, making it 100% compatible with Git, Obsidian, or backup tools.

---

## 🏗️ Architecture

Novel-TUI is organized using the **Go Standard Project Layout** coupled with **The Elm Architecture (TEA)**:

```text
novel-tui/
├── cmd/
│   └── novel-tui/               # CLI entrypoint and lifecycle bootstrapping
└── internal/
    ├── domain/                  # Pure entities (Chapter, Character, EditorMetrics)
    ├── repository/              # Filesystem storage with atomic rename writes
    ├── service/                 # Metric calculators (WPM, reading time) & workspace service
    └── ui/                      # Bubble Tea UI components
        ├── model/               # Root Model (layout calculation, global keybindings)
        ├── components/          # Sidebar, Editor (bubbles/textarea), StatusBar
        ├── messages/            # Decoupled TEA message definitions
        └── theme/               # Lip Gloss styles, colors, and borders
```

---

## 🧪 Testing

Run the automated test suite:

```bash
go test -v ./...
```

---

## 📄 License

MIT License © 2026 [Facundo Salvucci](https://github.com/SalvucciFacundo)
