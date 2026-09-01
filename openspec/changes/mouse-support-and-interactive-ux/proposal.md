# Proposal: Mouse Support and Interactive UX

## Intent
Enhance the overall user experience in `novel-tui` by integrating full mouse support. The goal is to allow users to interact naturally with UI elements using their mouse for actions like navigating the launcher, switching sidebar tabs, selecting chapters/lore, positioning the cursor in the editor, and scrolling through content—all while maintaining the existing keyboard shortcuts as the primary and fallback navigation method.

## Scope
This change introduces mouse interactions via Bubble Tea's `tea.WithMouseCellMotion()` and handles mouse events throughout the application's components:
1. **App Initialization:** Update `cmd/novel-tui/main.go` to inject mouse support into the `tea.Program`.
2. **Launcher:** Support clicking recent novels, clicking action menu items (`[c]`, `[n]`, etc.), and scrolling the list.
3. **Sidebar:** Support tab switching (Chapters vs Lore), clicking to select and load chapters, clicking to view lore, and mouse wheel scrolling.
4. **Editor:** Allow clicking to focus and position the cursor inside the text area, and scrolling text content.
5. **Non-breaking Constraint:** Keyboard bindings must continue working exactly as they do today.

## Affected Areas
- `cmd/novel-tui/main.go`: Add `tea.WithMouseCellMotion()` to `tea.NewProgram`.
- `internal/ui/components/launcher.go`: Add `tea.MouseMsg` handling (click coordinates detection, scroll up/down logic).
- `internal/ui/components/sidebar.go`: Add `tea.MouseMsg` handling for tabs and list elements.
- `internal/ui/components/editor.go`: Intercept and translate `tea.MouseMsg` to update cursor positioning and scroll offsets.
- `internal/ui/model/root.go`: Potential event routing adjustments if global mouse events need to be routed to specific focused components.

## Risks
- **Event Stealing/Overlap:** The TUI relies on coordinate mapping (`X`, `Y` from `tea.MouseMsg`) to determine which component was clicked. Incorrect mapping could lead to clicks triggering the wrong component (e.g., sidebar clicks triggering editor actions).
- **Editor Complexity:** Mapping terminal grid coordinates (cells) to actual text cursor positions inside the editor requires calculating text wrapping and scroll offsets. 
- **Terminal Compatibility:** Some terminal emulators do not fully support `WithMouseCellMotion()` or handle scrolling inconsistently, which could cause a disjointed experience for some users (though keyboard fallback mitigates this).

## Rollback
- Revert the `tea.WithMouseCellMotion()` initialization in `cmd/novel-tui/main.go`.
- Ignore or drop `tea.MouseMsg` handlers in the components. Since keyboard inputs are unchanged, safely ignoring mouse events restores exact previous behavior.

## Success Criteria
- [ ] Application starts with mouse capture enabled.
- [ ] Users can scroll through the launcher list and sidebar lists using the mouse wheel.
- [ ] Users can click a recent novel in the launcher to open it.
- [ ] Users can switch between "Chapters" and "Lore" tabs by clicking them.
- [ ] Clicking a chapter loads it into the editor; clicking a character card shows lore.
- [ ] Clicking within the editor positions the cursor approximately where clicked.
- [ ] All existing keyboard navigation works flawlessly without regression.

## Proposal question round
*To finalize this PRD and ensure the resulting design meets expectations, please review these assumptions and questions. You can answer, correct the framing, skip them, or approve to continue to the Specification phase.*

1. **Editor Coordinate Mapping:** Mapping raw terminal (X,Y) clicks to a specific character index inside a wrapped text editor can be complex. If exact character-level positioning proves difficult or inaccurate, is it acceptable for a click to simply focus the editor and move the cursor to the nearest line, or is strict precision required?
2. **Scroll Behavior:** For the editor and lists, how many lines should a single mouse wheel "tick" scroll? (e.g., 1 line for lists, 3 lines for the editor).
3. **Hover States:** Should components visually react to the mouse hovering over them (e.g., highlighting a menu item), or are we only handling explicit click/scroll events? (Bubble Tea's `WithMouseAllMotion` is needed for hover tracking, whereas `WithMouseCellMotion` only tracks clicks/drags. The requirements specify `WithMouseCellMotion`, so assume no hover effects for now).
4. **Action Menu Layout:** Since action menu items `[c]`, `[n]`, etc. in the launcher will be clickable, do they have fixed screen coordinates, or do we need to dynamically calculate their bounds on window resize?