# Proposal: novel-tui-layout

## Intent & Business Problem
Writers, particularly those working on light novels or fiction, currently lack a distraction-free, terminal-based tool that integrates chapter management, character tracking, and real-time writing metrics without heavy bloat. Existing graphical text editors can be distracting and consume significant system resources, while generic terminal editors require heavy configuration and plugins to serve as a focused fiction writing suite. There is an opportunity to provide a seamless, keyboard-centric workflow natively tailored to novel writing, directly in the terminal.

## Target Users and Situations
- **Light Novel / Fiction Writers**: Authors who want a distraction-free environment without leaving the terminal.
- **Keyboard-Centric Users**: Users who prefer fast navigation without a mouse, relying on clean keybindings (e.g., Tab, Ctrl+S) for productivity.
- **Local-First Writers**: Users who prefer owning their data in plain text/markdown on local disk for privacy, version control (git), and portability.

## Proposed Solution (Product Outcome)
Implement a robust multi-panel terminal UI using the Elm Architecture (Bubble Tea) and Lip Gloss for styling. The layout will consist of:
1. **Left Panel (Navigation)**: A list to manage, reorder, and select chapters, along with word counts per chapter.
2. **Central Panel (Editor)**: A distraction-free text editing area featuring real-time metrics (total word count, reading time, save status).
3. **Secondary Panel (Context)**: A toggleable or tabbed side panel for character cards (lore, background, role) and writing notes.
4. **Data Persistence**: Local file I/O layer saving chapters as Markdown/Text files and characters/metadata as JSON/YAML.

## Architecture & Implementation Overview
- **Framework**: `charmbracelet/bubbletea` for state management and event handling; `charmbracelet/lipgloss` for panel styling and responsive sizing.
- **State Model**: A unified `Model` struct that tracks the active focused panel, window dimensions, and delegates `Update()` messages to sub-models (ChapterList, Editor, ContextPanel).
- **Keybindings**: Global message interception for panel cycling (Tab/Shift-Tab), file operations (Ctrl+S, Ctrl+N), and command palettes (Ctrl+P).
- **Storage**: A local file repository interface abstracting os/fs operations to read/write chapter files and parse character metadata.

## Scope Boundaries & Non-Goals
- **Non-Goal**: This is not a generic code editor (no syntax highlighting for programming languages, no LSP integration).
- **Non-Goal**: No cloud synchronization or collaborative editing capabilities in the first slice; it remains strictly local.
- **Boundary**: Complex text formatting (bold/italics rendering in TUI) is out of scope for the layout slice. The editor will handle plain text/markdown source.
- **Boundary**: Deep Vim-emulation is not a primary goal for the initial MVP; standard intuitive bindings will be prioritized.

## Business Risks & Tradeoffs
- **Text Editing Complexity**: Implementing a performant, multiline text area with word-wrapping, scrolling, and cursor management natively in Bubble Tea can be highly complex. *Tradeoff*: We may need to leverage or extend existing components like `charmbracelet/bubbles/textarea` rather than building a custom editor from scratch.
- **Responsive Layouts**: Terminal resizing can break static layouts. We must ensure robust window resize message handling (`tea.WindowSizeMsg`) to dynamically adjust panel widths (e.g., fixed width for left/right panels, flex width for the center).

## Proposal Question Round

To finalize the PRD and ensure we are building the right slice, please review the following assumptions and answer the questions:

1. **Text Area Component**: Should we rely on `charmbracelet/bubbles/textarea` for the central editor, knowing it might have limitations with very large files, or do you have a specific text-editing engine in mind?
2. **Keybindings**: Do you want strictly standard UI bindings (Ctrl+S, arrows, Tab) or optional Vim motions for navigation (h,j,k,l)? 
3. **File Structure**: Should the tool enforce a specific directory structure (e.g., `src/chapters`, `data/characters.json`), or should it adapt to whatever is in the current working directory?
4. **Auto-save**: Is explicit saving (Ctrl+S) sufficient for the first slice, or is an auto-save loop (e.g., every 30 seconds or on focus lost) a hard requirement?

**Assumptions needing review:**
- The first slice will strictly target a single project directory opened in the terminal.
- We will prioritize standard keybindings over Vim motions to ensure accessibility for all writers.
- Markdown rendering (formatting) is not required inside the editor; plain text markdown source is sufficient.