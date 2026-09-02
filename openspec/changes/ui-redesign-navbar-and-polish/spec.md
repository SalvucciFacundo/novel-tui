# UI Redesign, Navbar, and Polish Specification

## Purpose

Define the functional and behavioral specifications for the redesigned user interface of `novel-tui`, including the top Navigation Bar (Navbar) with breadcrumbs and action pills, the enhanced Footer Status Bar with command badges and metrics, the 3-tab Sidebar system (Capítulos, Personajes, Notas), modal styling polish, editor viewport/cursor reset behavior, scoped keybindings (`Ctrl+N`), and return-to-home navigation (`Ctrl+H`).

---

## Requirements

### Requirement: Top Navigation Bar (Navbar) Layout and Breadcrumbs

The system MUST render a dedicated horizontal Navigation Bar at the top of the main workspace view, displaying contextual breadcrumbs that identify the active novel and currently loaded chapter.

#### Scenario: Displaying novel and chapter breadcrumbs
- GIVEN a novel named "Cien Años de Soledad" is opened in the editor view
- AND chapter "Capítulo 1: El Coronel Aureliano" is currently loaded
- WHEN the Top Navigation Bar is rendered
- THEN the system MUST display breadcrumbs showing the novel title and chapter title in the format `📖 Cien Años de Soledad › 📑 Capítulo 1: El Coronel Aureliano`
- AND the breadcrumb text MUST be visually distinct from action pills.

#### Scenario: Displaying breadcrumb fallback when no chapter is loaded
- GIVEN a novel is opened in the editor view
- AND no chapter has been selected or loaded yet
- WHEN the Top Navigation Bar is rendered
- THEN the breadcrumb MUST display `📖 <Novel Title> › 📑 Ningún capítulo seleccionado`.

---

### Requirement: Top Navigation Bar Interactive Action Pills

The system MUST render interactive action pills in the Top Navigation Bar and respond to mouse clicks on each pill.

#### Scenario: Action pill layout in Navbar
- GIVEN the Top Navigation Bar is rendered
- WHEN the action pill section is generated
- THEN the system MUST display the following action pills:
  - `[← Inicio (Ctrl+H)]`
  - `[1: Capítulos]`
  - `[2: Personajes]`
  - `[3: Notas]`
  - `[🤖 Asistente IA (Ctrl+A)]`.

#### Scenario: Clicking the Return Home pill
- GIVEN the Top Navigation Bar is active
- WHEN the user clicks with the left mouse button on `[← Inicio (Ctrl+H)]`
- THEN the system MUST transition the application view back to the Launcher Dashboard.

#### Scenario: Clicking Sidebar tab pills (Capítulos, Personajes, Notas)
- GIVEN the Top Navigation Bar is active and the editor workspace is visible
- WHEN the user clicks with the left mouse button on `[1: Capítulos]`
  - THEN the system MUST switch the active Sidebar tab to `TabChapters` (Tab 1)
- WHEN the user clicks with the left mouse button on `[2: Personajes]`
  - THEN the system MUST switch the active Sidebar tab to `TabCharacters` (Tab 2)
- WHEN the user clicks with the left mouse button on `[3: Notas]`
  - THEN the system MUST switch the active Sidebar tab to `TabNotes` (Tab 3).

#### Scenario: Clicking the AI Assistant pill
- GIVEN the Top Navigation Bar is active
- WHEN the user clicks with the left mouse button on `[🤖 Asistente IA (Ctrl+A)]`
- THEN the system MUST toggle the visibility of the AI Assistant Drawer
- AND focus the chat input when opening the drawer.

---

### Requirement: Bottom Footer Status Bar Command Badges and Metrics

The system MUST render an enhanced bottom status bar featuring styled command badges on the left side and real-time writing metrics with save status on the right side.

#### Scenario: Rendering left-aligned command badges
- GIVEN the workspace view is active
- WHEN the Statusbar is rendered
- THEN the left section MUST display styled command badges for core actions:
  - `[Ctrl+H: Inicio]`
  - `[Ctrl+A: IA]`
  - `[Ctrl+S: Guardar]`
  - `[Ctrl+N: Nuevo Cap]`
  - `[Tab: Panel]`.

#### Scenario: Rendering right-aligned metrics and save status indicator
- GIVEN a chapter with 1,250 words and 7,500 characters is loaded in the editor
- AND the average reading speed is 200 words per minute
- WHEN the editor content is saved with no pending changes
- THEN the right section of the Statusbar MUST display:
  - Word count: `1.250 palabras` (or localized formatting)
  - Character count: `7.500 caracteres`
  - Estimated reading time: `~6 min`
  - Save status badge: `[Guardado]` in green / affirmative style.

#### Scenario: Updating save status badge on buffer modification
- GIVEN a chapter is loaded and currently displays `[Guardado]`
- WHEN the user types or modifies text in the editor
- THEN the save status badge MUST immediately update to `[Modificado*]` in warning / modified accent color
- AND the word, character, and reading time metrics MUST update dynamically.

---

### Requirement: Sidebar 3-Tab Management System

The system MUST support a 3-tab navigation structure in the Sidebar: `Tab 1: Capítulos`, `Tab 2: Personajes`, and `Tab 3: Notas`, allowing users to switch between them via keyboard shortcuts (`1`, `2`, `3`), tab clicking, or Navbar pills.

#### Scenario: Tab 1: Capítulos (Chapters List)
- GIVEN `Tab 1: Capítulos` is active in the Sidebar
- WHEN rendered
- THEN the system MUST display the list of chapters with chapter index, title, and word count
- AND pressing `Enter` or double-clicking on a chapter MUST load that chapter into the editor
- AND pressing `n` while focused on the chapters list MUST open the New Chapter creation modal.

#### Scenario: Tab 2: Personajes (Character Lore Cards)
- GIVEN `Tab 2: Personajes` is active in the Sidebar
- WHEN characters exist in `personajes.json` (or `lore/`)
- THEN the system MUST render the list of characters
- AND selecting a character MUST render a detail card with Name, Role, Description, and Notes/Traits
- AND if no characters exist, the system MUST display an informative empty state with instructions on how to add characters.

#### Scenario: Tab 3: Notas (Novel Notes Viewer/Editor)
- GIVEN `Tab 3: Notas` is active in the Sidebar
- WHEN rendered
- THEN the system MUST load the content of `notas.txt` from the novel's root directory
- AND allow viewing and editing the notes directly or via a focused notes buffer
- AND modifications to `notas.txt` MUST be persisted when saving or switching tabs.

---

### Requirement: Modal Dialog Visual Polish

The system MUST render modal dialogs with consistent Lip Gloss background styling across all internal rows, borders, and padding, eliminating unstyled black or blank padding strips.

#### Scenario: Rendering modal with uniform background color
- GIVEN any modal dialog is displayed (e.g., New Chapter, New Novel, Settings, Confirmation)
- WHEN the modal is rendered on top of the workspace
- THEN all container margins, internal padding, input boxes, and button bars MUST inherit or explicitly apply the defined card background style
- AND the modal boundary MUST have a crisp, consistent border without unstyled dark lines or padding artefacts.

---

### Requirement: Editor Viewport and Cursor Reset on Chapter Load

The system MUST ensure that whenever a new or existing chapter is loaded into the editor, the textarea cursor and viewport are reset to the beginning of the text (Line 1, Column 1).

#### Scenario: Loading a chapter resets cursor and scroll position
- GIVEN the user was previously editing at Line 150 of Chapter 1
- WHEN the user selects and loads Chapter 2
- THEN the editor textarea MUST reset its cursor position to Line 1, Column 1
- AND the editor viewport MUST scroll to the top so that Line 1 is immediately visible at the top of the editor window.

---

### Requirement: Scoped Keybinding Isolation for Ctrl+N

The system MUST isolate the `Ctrl+N` keybinding so that chapter creation is triggered only when focus is on the Sidebar or the main Editor, and is ignored or passed as standard input when focused on the AI Chat drawer.

#### Scenario: Pressing Ctrl+N when focused on Editor or Sidebar
- GIVEN focus is on `FocusEditor` or `FocusSidebar`
- WHEN the user presses `Ctrl+N`
- THEN the system MUST open the "New Chapter" modal dialog.

#### Scenario: Pressing Ctrl+N when focused on AI Chat Drawer
- GIVEN the AI Assistant Chat Drawer is open and focused (`FocusChat`)
- WHEN the user presses `Ctrl+N`
- THEN the system MUST NOT open the "New Chapter" modal dialog
- AND the key event MUST either be handled by the chat textarea as standard cursor navigation/newline or ignored without side effects.

---

### Requirement: Return to Home Navigation

The system MUST provide clean navigation back to the Launcher Dashboard from the editor workspace via both keyboard shortcut `Ctrl+H` and clicking the `[← Inicio]` Navbar pill.

#### Scenario: Pressing Ctrl+H from any workspace focus state
- GIVEN the user is in the editor workspace in any focus state (`FocusEditor`, `FocusSidebar`, `FocusChat`, etc.)
- AND no modal dialog is actively capturing input
- WHEN the user presses `Ctrl+H`
- THEN the system MUST save any pending dirty buffer (or prompt if configured)
- AND transition the root application state to `ViewLauncher`
- AND refresh the recent novels list on the Launcher Dashboard.

#### Scenario: Clicking the [← Inicio] pill in the Navbar
- GIVEN the user clicks the `[← Inicio (Ctrl+H)]` pill with the mouse
- WHEN the click is processed
- THEN the system MUST perform the same clean transition back to `ViewLauncher`.
