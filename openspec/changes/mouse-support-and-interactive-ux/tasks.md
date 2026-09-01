## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~350-450 |
| 400-line budget risk | Medium |
| Chained PRs recommended | No |
| Suggested split | single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: stacked-to-main
400-line budget risk: Medium

## Implementation Tasks

- [x] 1. Enable `tea.WithMouseCellMotion()` in `cmd/novel-tui/main.go`. <!-- sdd-owner: implementation -->
- [x] 2. Implement mouse message routing and coordinate hit-testing in `internal/ui/model/root.go` (routing `tea.MouseMsg` to Modal, Launcher, Sidebar, or Editor based on spatial layout and ViewState). <!-- sdd-owner: implementation -->
- [x] 3. Implement mouse event handling in `internal/ui/components/launcher.go` (click on Action buttons [c], [n], [o], [l], [d], [q], click/double-click on recent novels list, mouse wheel scrolling). <!-- sdd-owner: implementation -->
- [x] 4. Implement mouse event handling in `internal/ui/components/sidebar.go` (click on [Chapters] and [Lore / Characters] tab headers, click on chapter row to select/open, click on character row, mouse wheel scrolling). <!-- sdd-owner: implementation -->
- [x] 5. Implement mouse event handling in `internal/ui/components/editor.go` (left click to focus editor and position cursor, mouse wheel scrolling 3 lines per tick). <!-- sdd-owner: implementation -->
- [x] 6. Implement mouse event handling in `internal/ui/components/modal.go` (click inside input focuses, clicks outside ignored safely). <!-- sdd-owner: implementation -->
- [x] 7. Write unit and component tests for mouse event dispatch, hit testing, and scrolling across all components (`internal/ui/components/components_test.go`, `internal/ui/model/root_test.go`). <!-- sdd-owner: implementation -->
- [x] 8. Perform post-apply verification and ensure all keyboard shortcuts remain intact without regression. <!-- sdd-owner: implementation -->
