# Software Architecture Design: UI Redesign, Navbar, and Polish

## 1. System Overview

This redesign enhances the `novel-tui` workspace by introducing a top `Navbar` for breadcrumbs and navigation, refining the `Statusbar` for metrics and global commands, adding a `Notas` tab to the `Sidebar`, and fixing UI/UX defects (modal styling, editor cursor positioning, and scoped keybindings). 

The changes are contained within the presentation layer (`internal/ui/components/` and `internal/ui/model/root.go`), maintaining the existing BubbleTea architecture pattern (Model-Update-View).

## 2. Component Design

### 2.1 Top Navigation Bar (`internal/ui/components/navbar.go`)
- **Model**: A new `Navbar` struct holding `NovelTitle`, `ChapterTitle`, `Width`, and interactive bounding box state for mouse clicks.
- **View (`View() string`)**:
  - Rendered using `lipgloss.JoinHorizontal(lipgloss.SpaceBetween, leftSection, rightSection)`.
  - **Left**: `[← Inicio (Ctrl+H)]` pill + `📖 <NovelTitle> › 📑 <ChapterTitle>`.
  - **Right**: Action pills `[1: Capítulos]`, `[2: Personajes]`, `[3: Notas]`, `[🤖 Asistente IA (Ctrl+A)]`.
- **Update (`Update(msg tea.Msg) (Model, tea.Cmd)`)**:
  - Listens for `tea.MouseMsg`.
  - Detects mouse clicks on calculated X-coordinates for the action pills.
  - Emits specific commands on click (e.g., `ReturnHomeMsg`, `ChangeSidebarTabMsg(tabIndex)`, `ToggleAIChatMsg`).
- **Integration**: Placed at the top of the workspace layout in `root.go` / `workspace.go`.

### 2.2 Status Bar (`internal/ui/components/statusbar.go`)
- **Model Update**: Introduce fields for metrics (`Words`, `Chars`, `ReadingTime`) and `IsSaved` boolean.
- **View Update**:
  - **Left**: Command badges created via `lipgloss.NewStyle().Background(...)`. Badges: `[Ctrl+H: Inicio]`, `[Ctrl+A: IA]`, `[Ctrl+S: Guardar]`, `[Ctrl+N: Nuevo Cap]`, `[Tab: Panel]`.
  - **Right**: Format metrics (`{Words} palabras | {Chars} caracteres | ~{Time} min`). 
  - **Save Badge**: Appended to the far right. If `IsSaved` is true, render `[Guardado]` in a green/success style. If false, render `[Modificado*]` in an orange/warning style.

### 2.3 Sidebar (`internal/ui/components/sidebar.go`)
- **State Changes**:
  - Add `activeTab int` with constants `TabChapters (0)`, `TabCharacters (1)`, `TabNotes (2)`.
  - Add `notesTextarea textarea.Model` to handle the `notas.txt` editor.
- **Data Handling (`notas.txt`)**:
  - On novel load, check if `notas.txt` exists in the novel's root directory.
  - Read contents into `notesTextarea`.
  - Provide a command to save `notas.txt` to disk when switching tabs or issuing a global save.
- **View**:
  - Render tab headers indicating the active tab.
  - Conditionally render the body: list of chapters, list of characters, or the notes textarea.

### 2.4 Modal Visual Polish (`internal/ui/components/modal.go`)
- **Defect**: "Black strips" appear because inner containers do not inherit the background color of the modal card boundary, or padding does not fill the width.
- **Fix**:
  - Ensure all inner elements (title, body, buttons) use `.Background(modalBgColor)`.
  - Use `lipgloss.Place(...)` or explicit `.Width(innerMaxWidth)` to ensure child elements stretch to the full modal width without leaving empty terminal background cells.

### 2.5 Editor Cursor Reset (`internal/ui/components/editor.go`)
- **Update Logic**:
  - Intercept the message emitted when a chapter is selected/loaded (e.g., `ChapterLoadedMsg` or `EditorLoadMsg`).
  - Upon receiving this message, explicitly call `m.textarea.SetCursor(0)` (or `CursorStart()`, depending on the underlying `bubbles/textarea` version).
  - Reset vertical scroll offset/viewport to `0` so the editor always renders from Line 1, Column 1.

### 2.6 Root/Workspace Routing & Keybindings (`internal/ui/model/root.go`)
- **Layout Construction**:
  - Adjust the height available to `Editor`, `Sidebar`, and `Chat` by subtracting the heights of the `Navbar` (1-2 lines) and `Statusbar` (1-2 lines) from `msg.Height`.
  - Render sequence: `lipgloss.JoinVertical(lipgloss.Left, navbar.View(), workspace.View(), statusbar.View())`.
- **Keybinding Interception (`Update`)**:
  - `Ctrl+H`: If received anywhere, dispatch a state change to `ViewLauncher`.
  - `Ctrl+N`: Add a focus guard. If `m.focus == FocusChat`, ignore `Ctrl+N` (do not spawn the New Chapter modal). Only trigger chapter creation if `m.focus == FocusEditor` or `m.focus == FocusSidebar`.

## 3. Data Flow

1. **Navigation**: User clicks a Navbar pill -> `Navbar` emits `ChangeSidebarTabMsg(2)` -> `Root` routes message to `Sidebar` -> `Sidebar` sets `activeTab = TabNotes` -> UI rerenders.
2. **Metrics & Save State**: User types in `Editor` -> text changes -> `Editor` calculates new word/char count -> Emits `EditorMetricsMsg` & `BufferDirtyMsg` -> `Root` routes to `Statusbar` -> `Statusbar` updates and displays `[Modificado*]`.
3. **Saving Notes**: `Sidebar` in `TabNotes` receives text input -> User hits `Ctrl+S` -> `Root` broadcasts save -> `Sidebar` flushes `notas.txt` to disk.

## 4. Risks and Mitigation

- **Terminal Mouse Support**: Clickable pills require `tea.EnableMouseCellMotion` or similar. We must ensure the program initializes mouse support in `main.go`.
- **Layout Calculation**: Hardcoding heights can cause layout collapse on small terminal windows. **Mitigation**: Use dynamic height calculation (`contentHeight = termHeight - navbarHeight - statusbarHeight`) before passing height to the Editor and Sidebar components.
- **Focus Conflicts**: Giving the Sidebar its own Textarea (for Notes) might conflict with global text input captures. **Mitigation**: Strictly manage the `m.focus` enum to determine which Textarea component receives `tea.KeyMsg` during the `Update` loop.

## 5. Implementation Phases
1. Scaffold `Navbar` component, adjust `Root` layout, and test sizing.
2. Update `Statusbar` to display left badges and right metrics dynamically.
3. Refactor `Sidebar` to support 3 tabs and wire up the `notas.txt` textarea.
4. Implement the `Ctrl+N` focus guard, `Ctrl+H` handler, and Editor cursor reset.
5. Apply lipgloss background uniform styles to `modal.go`.
