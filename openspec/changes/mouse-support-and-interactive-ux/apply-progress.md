# Apply Progress: Mouse Support and Interactive UX

## Completed Tasks
- [x] 1. Enable `tea.WithMouseCellMotion()` in `cmd/novel-tui/main.go`.
- [x] 2. Implement mouse message routing and coordinate hit-testing in `internal/ui/model/root.go` (routing `tea.MouseMsg` to Modal, Launcher, Sidebar, or Editor based on spatial layout and ViewState).
- [x] 3. Implement mouse event handling in `internal/ui/components/launcher.go` (click on Action buttons [c], [n], [o], [l], [d], [q], click/double-click on recent novels list, mouse wheel scrolling).
- [x] 4. Implement mouse event handling in `internal/ui/components/sidebar.go` (click on [Chapters] and [Lore / Characters] tab headers, click on chapter row to select/open, click on character row, mouse wheel scrolling).
- [x] 5. Implement mouse event handling in `internal/ui/components/editor.go` (left click to focus editor and position cursor, mouse wheel scrolling 3 lines per tick).
- [x] 6. Implement mouse event handling in `internal/ui/components/modal.go` (click inside input focuses, clicks outside ignored safely).
- [x] 7. Write unit and component tests for mouse event dispatch, hit testing, and scrolling across all components (`internal/ui/components/components_test.go`, `internal/ui/model/root_test.go`).
- [x] 8. Perform post-apply verification and ensure all keyboard shortcuts remain intact without regression.

## Files Changed
- `cmd/novel-tui/main.go`
- `internal/ui/components/launcher.go`
- `internal/ui/components/sidebar.go`
- `internal/ui/components/editor.go`
- `internal/ui/components/modal.go`
- `internal/ui/model/root.go`
- `internal/ui/components/components_test.go`
- `internal/ui/model/root_test.go`
- `openspec/changes/mouse-support-and-interactive-ux/tasks.md`

## Verification Evidence
- `go test -v ./...`: All unit and component tests pass across all packages (domain, repository, service, components, model).
- `go build ./cmd/novel-tui`: Binary compiles cleanly without warnings or errors.
- Both mouse interactions and keyboard shortcuts verified with automated tests.
