# novel-tui-layout Specification

## Purpose

Define the functional and interface requirements for the multi-panel Terminal User Interface (TUI) layout of `novel-tui`. The system provides fiction and light novel writers with a distraction-free, keyboard-centric environment composed of a multi-tab sidebar (Chapters and Lore/Characters), a central text editor, and a bottom status bar showing live writing metrics and persistence indicators.

## Requirements

### Requirement: Multi-Panel Layout and Sizing

The system MUST render a responsive three-section terminal layout consisting of a Left Sidebar, a Center Editor, and a Bottom Status Bar, dynamically adjusting to terminal window dimensions.

#### Scenario: Standard terminal window rendering
- GIVEN a terminal window with width >= 80 columns and height >= 24 rows
- WHEN the application initializes or receives a window resize message (`tea.WindowSizeMsg`)
- THEN the layout MUST allocate:
  - Left Sidebar: fixed width between 24 and 32 columns (default 28 columns), height = window height - status bar height (1 row) - border offsets
  - Center Editor: flex width = total window width - Left Sidebar width - border offsets, height = window height - status bar height (1 row) - border offsets
  - Bottom Status Bar: full window width (100%), fixed height of 1 row at the bottom of the screen.

#### Scenario: Terminal window too small
- GIVEN a terminal window with width < 60 columns or height < 15 rows
- WHEN a window resize message is processed
- THEN the system MUST render a clear warning view instructing the user to resize the terminal instead of rendering a broken layout.

---

### Requirement: Navigation and Focus State Machine

The system MUST maintain an explicit focus state machine that tracks the currently active panel and provides visual cues and keyboard navigation between panels.

#### Scenario: Cycling panel focus forward and backward
- GIVEN the application is running with active focus on the Left Sidebar
- WHEN the user presses `Tab`
- THEN focus MUST transition to the Center Editor.
- WHEN the user presses `Tab` again
- THEN focus MUST cycle back to the Left Sidebar (or the next focusable interactive panel).
- WHEN the user presses `Shift+Tab`
- THEN focus MUST cycle in the reverse order.

#### Scenario: Visual focus indicator
- GIVEN a panel is currently focused
- WHEN the UI renders via Lip Gloss styling
- THEN the focused panel's border MUST render with an active highlight style (e.g., distinct accent color/border), and unfocused panels MUST render with a muted/dim border style.

---

### Requirement: Sidebar Tab Switching (Chapters vs Lore/Characters)

The Left Sidebar MUST support switching between two distinct contextual views: Chapters list and Lore/Characters list.

#### Scenario: Toggling between sidebar tabs
- GIVEN the Left Sidebar has focus
- WHEN the user presses `[` or `]` (or tab switch hotkeys `h`/`l` when in sidebar header mode)
- THEN the active view MUST toggle between `Chapters` and `Lore/Characters`.
- AND the sidebar header MUST visually indicate which tab is currently active.

---

### Requirement: Chapter Management in Sidebar

The system MUST display existing chapters, allow selection, indicate chapter metrics, and support creating new chapters.

#### Scenario: Chapter list display and selection
- GIVEN the Left Sidebar is active on the `Chapters` tab with one or more chapters loaded
- WHEN the user navigates with `Up`/`Down` (or `k`/`j` when not editing text)
- THEN the selected item indicator MUST update.
- WHEN the user presses `Enter` on a selected chapter
- THEN the Center Editor MUST load the selected chapter's content into the editing buffer.

#### Scenario: Creating a new chapter
- GIVEN the user presses `Ctrl+N` (or `n` while focused on Chapter list)
- WHEN a new chapter name is submitted or generated
- THEN a new chapter entry MUST be added to the chapter list.
- AND the new chapter MUST be loaded into the Center Editor with focus transferred to the editor.

---

### Requirement: Character and Lore Panel

The system MUST provide a character and lore viewer within the sidebar when the `Lore/Characters` tab is active.

#### Scenario: Viewing character list and card details
- GIVEN the sidebar is active on the `Lore/Characters` tab
- WHEN the user navigates the character list with `Up`/`Down`
- THEN the panel MUST display the selected character's summary card (name, role, description, and key notes).
- AND if character list is empty, a friendly placeholder with instructions to add characters MUST be displayed.

---

### Requirement: Text Editor Component Behavior

The Center Editor MUST wrap `charmbracelet/bubbles/textarea` (or equivalent multiline editor component) providing word wrapping, cursor navigation, optional line numbers, and live metric tracking.

#### Scenario: Text editing and word wrapping
- GIVEN the Center Editor has focus
- WHEN the user types text into the editor
- THEN characters MUST appear at the cursor position.
- AND long lines MUST soft-wrap at word boundaries to fit the editor's allocated width.

#### Scenario: Live metric event propagation
- GIVEN the user modifies the text buffer in the editor
- WHEN any insertion or deletion occurs
- THEN the editor MUST immediately recalculate word count, character count, and mark the buffer state as modified (`isDirty = true`).

#### Scenario: Line number display toggle
- GIVEN the editor is rendered
- WHEN line numbers are enabled in configuration/state
- THEN line numbers MUST render in a left gutter with muted styling without interfering with text input or copy buffers.

---

### Requirement: Status Bar Metrics and State Indicators

The Bottom Status Bar MUST display real-time session information, including current chapter title, word count, character count, estimated reading time, and buffer save status.

#### Scenario: Displaying status bar metrics
- GIVEN an active chapter named "Chapter 1: The Beginning" loaded with 500 words and 2,500 characters
- WHEN the status bar is rendered
- THEN it MUST display:
  - Chapter title: "Chapter 1: The Beginning"
  - Word count: "500 words"
  - Character count: "2500 chars"
  - Reading time: "~2 min read" (calculated at 200–250 words per minute)
  - Save indicator: `[Saved]` when clean or `[Modified*]` when `isDirty` is true.

---

### Requirement: Local File Persistence and Directory Structure

The system MUST interact with the local filesystem through a predictable directory layout, saving chapter text as Markdown files and character lore as JSON.

#### Scenario: Saving active chapter to disk
- GIVEN the editor buffer is modified (`isDirty = true`)
- WHEN the user triggers save via `Ctrl+S`
- THEN the system MUST write the text buffer content to `<project_root>/chapters/<chapter_id_or_slug>.md`.
- AND the save indicator in the status bar MUST update to `[Saved]`.
- AND `isDirty` MUST become `false`.

#### Scenario: Loading project workspace on startup
- GIVEN a project directory containing `chapters/` and `characters.json`
- WHEN `novel-tui` launches in that directory
- THEN the system MUST scan and parse all `.md` files in `chapters/` into the chapter repository.
- AND parse `characters.json` into the lore repository.
- AND if `chapters/` directory does not exist, the system MUST create it automatically.
