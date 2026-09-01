# Mouse Support and Interactive UX Specification

## Purpose

Define the functional and behavioral specifications for mouse interactions in `novel-tui`, including terminal mouse event capture, launcher action button clicks, recent novel selection and opening, sidebar tab and item interactions, editor focusing and scrolling, and strict non-regression of existing keyboard controls.

## Requirements

### Requirement: Application Lifecycle Mouse Event Tracking

The system MUST initialize the terminal program with mouse cell motion tracking enabled alongside the fullscreen alternate screen buffer, ensuring clean teardown and restoration upon application termination.

#### Scenario: Application startup with mouse cell motion
- GIVEN the `novel-tui` application binary is executed from a terminal
- WHEN `tea.NewProgram` is initialized in `cmd/novel-tui/main.go`
- THEN the program options MUST include `tea.WithMouseCellMotion()`
- AND the program options MUST include `tea.WithAltScreen()`
- AND terminal mouse events (`tea.MouseMsg`) MUST be delivered to the root UI model during runtime.

#### Scenario: Clean application exit with mouse teardown
- GIVEN the application is running with active mouse cell motion tracking
- WHEN the user exits the application via `Ctrl+C`, `q`, or application quit commands
- THEN the system MUST disable mouse cell motion capture in the terminal
- AND restore the terminal buffer and cursor without lingering mouse tracking mode artifacts.

---

### Requirement: Launcher Dashboard Mouse Interactions

The system MUST allow users to interact with all Launcher UI elements using mouse clicks and mouse wheel scrolling, including selecting items, activating action buttons, opening novels, and scrolling the recent novels list.

#### Scenario: Clicking an action menu button
- GIVEN the Launcher Dashboard is visible and active
- WHEN the user clicks with the left mouse button on any action menu button bounds:
  - `[c]` Continue Last Project
  - `[n]` New Novel
  - `[o]` Open Other Folder
  - `[l]` Configure LLM / AI
  - `[d]` Set Root Directory
  - `[q]` Quit
- THEN the system MUST trigger the corresponding action immediately:
  - Clicking `[c]` MUST trigger continuing the last accessed project
  - Clicking `[n]` MUST open the New Novel modal dialog
  - Clicking `[o]` MUST trigger opening the other folder selector
  - Clicking `[l]` MUST navigate to the LLM Configuration view
  - Clicking `[d]` MUST open the Root Directory settings modal
  - Clicking `[q]` MUST emit a quit message to terminate the application.

#### Scenario: Single-clicking a recent novel item
- GIVEN the Launcher Dashboard contains a list of recent novels
- WHEN the user clicks with the left mouse button on a novel row in the recent novels list
- THEN the system MUST update the active selection index to highlight the clicked novel
- AND the UI MUST re-render reflecting the new selection.

#### Scenario: Double-clicking a recent novel item
- GIVEN the Launcher Dashboard contains a list of recent novels
- WHEN the user rapidly double-clicks (within standard double-click threshold or consecutive clicks on the same item) on a novel row
- THEN the system MUST load the targeted novel project
- AND transition the view state from Launcher to Editor view.

#### Scenario: Mouse wheel scrolling on recent novels list
- GIVEN the Launcher Dashboard has more recent novel items than the visible list viewport height
- WHEN the user scrolls the mouse wheel down over the recent novels list area
- THEN the system MUST scroll the list selection or viewport downward by 1 item per wheel tick (clamped to the last item)
- WHEN the user scrolls the mouse wheel up over the recent novels list area
- THEN the system MUST scroll the list selection or viewport upward by 1 item per wheel tick (clamped to the first item).

---

### Requirement: Sidebar Mouse Navigation and Item Selection

The system MUST allow users in the Editor view to switch sidebar tabs, select and load chapters, display character lore cards, and scroll through sidebar lists using the mouse.

#### Scenario: Clicking the Chapters tab
- GIVEN the Editor view is active and the Sidebar is in `TabLore` mode
- WHEN the user clicks with the left mouse button on the `[Chapters]` tab header in the sidebar
- THEN the system MUST set `ActiveTab` to `TabChapters`
- AND render the chapters list in the sidebar.

#### Scenario: Clicking the Lore / Characters tab
- GIVEN the Editor view is active and the Sidebar is in `TabChapters` mode
- WHEN the user clicks with the left mouse button on the `[Lore / Characters]` tab header in the sidebar
- THEN the system MUST set `ActiveTab` to `TabLore`
- AND render the character and lore list in the sidebar.

#### Scenario: Clicking a chapter in the chapters list
- GIVEN the Sidebar is displaying the `TabChapters` list
- WHEN the user clicks with the left mouse button on a specific chapter entry in the list
- THEN the system MUST select that chapter
- AND load the chapter's text content into the Center Editor
- AND set the active focus or editor state to the loaded chapter.

#### Scenario: Clicking a character in the lore list
- GIVEN the Sidebar is displaying the `TabLore` list
- WHEN the user clicks with the left mouse button on a specific character or lore entry
- THEN the system MUST select that character
- AND display the character's details or bio card in the lore view pane.

#### Scenario: Mouse wheel scrolling in Sidebar lists
- GIVEN the active sidebar list (Chapters or Lore) has overflow items exceeding visible panel height
- WHEN the user scrolls the mouse wheel down over the Sidebar bounds
- THEN the system MUST scroll the active sidebar list down by 1 item per wheel tick
- WHEN the user scrolls the mouse wheel up over the Sidebar bounds
- THEN the system MUST scroll the active sidebar list up by 1 item per wheel tick.

---

### Requirement: Editor Mouse Focus and Content Scrolling

The system MUST handle mouse events inside the Editor viewport to focus the text editor on click and scroll the document text content smoothly via the mouse wheel.

#### Scenario: Clicking inside the Editor panel
- GIVEN the Editor view is active and focus is currently on the Sidebar or Statusbar
- WHEN the user clicks with the left mouse button within the rectangular bounds of the Center Editor
- THEN the system MUST transfer keyboard focus to the Editor (`FocusEditor`)
- AND activate the cursor inside the textarea.

#### Scenario: Mouse wheel scrolling inside Editor
- GIVEN the Editor contains document text exceeding the visible textarea height
- WHEN the user scrolls the mouse wheel down over the Editor viewport
- THEN the system MUST scroll the textarea content downward by 3 lines per wheel tick
- WHEN the user scrolls the mouse wheel up over the Editor viewport
- THEN the system MUST scroll the textarea content upward by 3 lines per wheel tick.

---

### Requirement: Non-Regression of Keyboard Navigation and Hotkeys

The system MUST ensure that all existing keyboard shortcuts, modal hotkeys, and focus switching bindings remain fully operational and unhindered by mouse event processing.

#### Scenario: Launcher keyboard navigation remains intact
- GIVEN the Launcher Dashboard is visible
- WHEN the user presses `Up`/`Down` or `k`/`j` keys
- THEN the novel list selection MUST navigate up and down exactly as before
- AND pressing `Enter` on a selected novel MUST open it
- AND pressing single-key shortcuts (`c`, `n`, `o`, `l`, `d`, `q`) MUST execute their respective actions.

#### Scenario: Sidebar and Editor keyboard shortcuts remain intact
- GIVEN the Editor view is active
- WHEN the user presses `Tab` or `Shift+Tab`
- THEN focus MUST cycle predictably between Sidebar and Editor
- AND pressing `[` / `]` or `h` / `l` in the Sidebar MUST switch tabs between Chapters and Lore
- AND pressing `Ctrl+S` MUST save the active chapter
- AND pressing `Ctrl+N` MUST open the New Chapter modal dialog
- AND pressing `Ctrl+H` MUST return to the Launcher Dashboard.
