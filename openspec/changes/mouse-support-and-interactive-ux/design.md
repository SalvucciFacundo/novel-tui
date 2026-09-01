# Architecture Design: Mouse Support and Interactive UX

## 1. Overview
The `novel-tui` project is expanding its terminal user interface to support mouse interactions (clicks and scrolling) via Bubble Tea. This change adds intuitive mouse support for the launcher, sidebar, and editor, without regressing any existing keyboard navigation.

## 2. Program Lifecycle Options
In `cmd/novel-tui/main.go`, the application initializes the `tea.Program`. Currently, it uses `tea.WithAltScreen()`. To enable mouse support, we will add `tea.WithMouseCellMotion()`:
```go
p := tea.NewProgram(rootModel, tea.WithAltScreen(), tea.WithMouseCellMotion())
```
- `tea.WithMouseCellMotion()` tracks mouse clicks, wheel scrolling, and drag events at the cell level.
- Clean application exit automatically handles mouse teardown because Bubble Tea manages terminal mode restoration.

## 3. Root Model Mouse Event Routing
The `RootModel` (`internal/ui/model/root.go`) acts as the top-level event dispatcher. When it receives a `tea.MouseMsg`, it must route it based on the current `viewState`, modal visibility, and spatial layout.

### 3.1 Global Interception
If the modal dialog is active (`m.modal.Active`), all `tea.MouseMsg` events must be passed *only* to the modal, bypassing the underlying view. Currently, the modal handles centering, so if clicks fall outside the modal's centered box, they should be safely ignored.

### 3.2 View-State Dispatch
- **`ViewStateLauncher`**: Pass the message to `m.launcher`.
- **`ViewStateLLMConfig`**: Pass the message to `m.llmConfig`.
- **`ViewStateEditor`**: The screen is split between the Sidebar (left) and the Editor (right). The Statusbar is at the bottom.
  - **Hit-Testing**:
    - If `Y >= m.height - 1`: Route to `statusbar`.
    - If `X < m.sidebar.width`: Route to `sidebar`.
    - If `X >= m.sidebar.width`: Translate coordinates to local (`msg.X -= m.sidebar.width`) and route to `editor`.

*Design Decision*: Translating coordinates before passing them to child components (like `editor`) ensures components can calculate hit-testing relative to `(0,0)`, keeping them decoupled from their global layout position.

## 4. Component Hit-Testing & Coordinate Calculations

### 4.1 LauncherModel
The `LauncherModel` renders its content dynamically centered using `lipgloss.Place`. Because the physical coordinates vary based on `m.width` and `m.height`, the launcher must calculate the absolute `(X, Y)` bounds of its interactable elements during (or immediately before) `View()` and store them, or calculate them mathematically on `Update()`.

- **Menu Action Buttons (Left Box)**: 
  - Calculates its box's `StartX` and `StartY` relative to the centered layout.
  - Buttons (`[c]`, `[n]`, `[o]`, `[l]`, `[d]`, `[q]`) are stacked vertically. A left-click (`tea.MouseLeft`) inside the box bounds, specifically on a Y-coordinate mapping to a button row, emits the equivalent action message.
- **Recent Novels List (Right Box)**:
  - Scroll wheel (`tea.MouseWheelUp` / `tea.MouseWheelDown`) over this bounds triggers the same index mutation as `keys.Up` and `keys.Down`.
  - Left-clicking a row (each novel takes ~2 lines) calculates the selected index: `novelIndex = (clickY - listStartY) / 2`. A single click selects; a double click (or clicking an already selected item) opens the novel.

### 4.2 SidebarModel
The Sidebar uses fixed internal layouts with an adjustable overall height.
- **Tab Header**: The top 2 lines (typically `Y == 0` or `Y == 1`) hold the tabs. 
  - `[Chapters]` tab starts near `X = 2`.
  - `[Lore / Characters]` tab follows.
  - A click on these bounding regions updates `ActiveTab` and re-renders the list.
- **List Items**:
  - Starts below the header (e.g., `Y = 3`).
  - Item clicks determine the target index using `selectedIndex = Y - listStartY + scrollOffset`.
  - Clicking an item emits `messages.ChapterSelectedMsg` (or lore equivalent).
- **Scrolling**: Mouse wheel events (`tea.MouseWheelUp` / `tea.MouseWheelDown`) increment or decrement the list viewport's `scrollOffset` by 1 item, clamped to the list bounds.

### 4.3 EditorModel
The Editor receives coordinates relative to its own top-left corner (`0,0`).
- **Focus Activation**: A `tea.MouseLeft` anywhere inside its bounds emits `FocusMsg{Target: FocusEditor}`, transferring keyboard control back to the editor.
- **Scrolling**: `tea.MouseWheelUp` and `tea.MouseWheelDown` adjust the vertical scroll offset of the viewport (typically 3 lines per tick for faster navigation in long texts).
- **Cursor Positioning (Future/Best Effort)**: Maps the clicked `(X, Y)` to the actual text string index based on the wrapped lines array. If exact coordinate mapping is imprecise, it snaps to the closest line or simply focuses the editor.

## 5. Safety & Edge Cases

1. **Out-of-Bounds Clicks**: 
   - All hit-testing must clamp or ignore coordinates that fall in empty padding or borders (e.g., `lipgloss` margins).
   - If a user clicks outside the list bounds but inside the parent layout, it should do nothing.
2. **Resizing Boundary Recalculation**: 
   - Because `lipgloss` layouts dynamically adapt to `tea.WindowSizeMsg`, static coordinate assumptions will fail. All hit boxes must be derived from `m.width` and `m.height` directly.
3. **Double Click Tracking**:
   - `tea.MouseMsg` does not inherently emit "double click". We must track the `time.Time` and `(X, Y)` of the last left click. If a second click occurs within ~300ms at the same `(X, Y)` (or same component index), it is treated as a double-click.
4. **Coordinate Drift**:
   - Borders and paddings added by `lipgloss.NewStyle().Border(...)` shift the inner content by 1-2 cells. Hit-testing offsets must account for these decorative borders.

## 6. Testing & Rollback
- **Validation**: Strict TDD can be applied where view layout functions are pure. Coordinate mapping functions can be tested via standard table-driven unit tests.
- **Rollback**: Removing `tea.WithMouseCellMotion()` from `main.go` acts as a hard kill switch, reverting the application entirely to keyboard-only mode without breaking any component logic.