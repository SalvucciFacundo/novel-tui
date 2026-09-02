# Proposal: UI Redesign, Navbar, and Polish

## 1. Intent and Business Problem
The current UI relies on sidebar tabs and a bottom status bar that underutilizes space, lacking clear top-level orientation. Furthermore, there are UX defects: the editor view sometimes clips the first line on load, modal components show unstyled black strips due to padding/width mismatches, and global keybindings like `Ctrl+N` fire unexpectedly (e.g., opening a new chapter modal while typing in the AI chat drawer). 

This redesign introduces a dedicated top Navigation Bar (Navbar) for clear breadcrumbs and global actions, revamps the Footer to display styled command badges, adds a global Notes tab to the sidebar, and polishes the structural UI/UX defects for a seamless writing environment.

## 2. Scope and Product Outcome
- **Top Navigation Bar (Navbar)**:
  - Display context breadcrumbs: `[← Inicio (Ctrl+H)]` `📖 <Novel Title>` > `📑 <Current Chapter Title>`.
  - Provide interactive clickable pills: `[1: Capítulos]`, `[2: Personajes]`, `[3: Notas]`, and `[🤖 Asistente IA: Ctrl+A]`.
- **Bottom Footer Bar (Statusbar)**:
  - Replace the long empty space with clear, styled command badges: `[Ctrl+H: Inicio]`, `[Ctrl+A: IA]`, `[Ctrl+S: Guardar]`, `[Ctrl+N: Nuevo Cap]`, and `[Tab: Panel]`.
  - Right-align live writing metrics: words, characters, reading time, and the current save state (`[Guardado]` / `[Modificado*]`).
- **Sidebar (3 Tabs)**:
  - Expand to three tabs: `Tab 1: Capítulos` (chapter list with counts), `Tab 2: Personajes` (lore bios), and `Tab 3: Notas` (a new viewer/editor for `notas.txt`).
- **Modal UI Polish**:
  - Standardize background rendering in `modal.go` so internal elements and padding match the card background, eliminating unstyled "black bars".
- **Editor Viewport Fix**:
  - Reset the textarea viewport/cursor upon loading a chapter so that Line 1 is always visible and never inadvertently scrolled out of view.
- **Keybinding Scope (Ctrl+N)**:
  - Isolate the `Ctrl+N` keybinding so it creates a new chapter only when focused on the editor or sidebar, preventing interference when the AI Chat Drawer has focus.
- **Return to Home**:
  - Implement reliable handling for `Ctrl+H` and clicks on `[← Inicio]` to transition back cleanly to the Launcher Dashboard.

## 3. Affected Areas
- **New Component**: `internal/ui/components/navbar.go` (handles top bar rendering and clicks).
- **Statusbar**: `internal/ui/components/statusbar.go` (layout rework and badge styling).
- **Sidebar**: `internal/ui/components/sidebar.go` (addition of `TabNotas`, integration with `notas.txt`).
- **Editor**: `internal/ui/components/editor.go` (textarea cursor/viewport reset on `ChapterSelectedMsg`).
- **Modal**: `internal/ui/components/modal.go` (lipgloss styles and padding adjustments).
- **Root Model**: `internal/ui/model/root.go` (layout composition for the navbar, `Ctrl+H` handler, `Ctrl+N` focus guards).

## 4. Risks and Rollback
- **Risks**: 
  - Inserting a horizontal Navbar reduces the vertical space for the editor by 1-2 lines. The terminal `MinHeight` constraint might need a slight adjustment to prevent layout collapse.
  - Adding a `notas.txt` editor inside the sidebar adds state management for another text area, requiring careful handling to not conflict with the main editor's focus.
- **Rollback**: Changes are contained entirely within the UI layer (BubbleTea models) and do not permanently modify novel data structures. A standard Git revert safely unwinds the redesign without data loss.

## 5. Success Criteria
- The Navbar appears at the top of the editor view with accurate breadcrumbs and 4 clickable pills.
- The Statusbar correctly displays all 5 command badges and right-aligned metrics.
- The Sidebar properly switches between Capítulos, Personajes, and Notas.
- `Ctrl+N` types normally in the chat drawer without triggering a "New Chapter" modal.
- Changing chapters instantly shows the first line in the editor without manual scrolling.
- Modals render cleanly with uniform card backgrounds.
- `Ctrl+H` correctly unmounts the editor and returns to the Launcher.

## Proposal Question Round
*(If using Interactive Mode, please answer or correct the following assumptions before we proceed to specifications):*

1. **Product Behavior**: Does clicking a Navbar pill (e.g., `[1: Capítulos]`) just switch the active tab in the left Sidebar, or does it trigger a full-screen view? (Assumption: it just switches the Sidebar tab).
2. **Notes Storage**: For `Tab 3: Notas`, is `notas.txt` a global file per novel, or a file per chapter? (Assumption: global per novel, stored in the novel's root directory).
3. **Empty States**: If no chapter is selected yet, should the breadcrumb just say `📑 No Chapter Selected`? (Assumption: Yes, matching current default behavior).
