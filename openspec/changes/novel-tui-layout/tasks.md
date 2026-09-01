## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 1200-1800 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1: Setup, Domain & Repositories (internal/domain, internal/repository) → PR 2: Services, Theme & UI Sub-Components (internal/service, internal/ui/theme, internal/ui/components) → PR 3: Root Model, Main Entry Point & Unit Tests (internal/ui/model, cmd/novel-tui, test suite) |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

---

## Tasks

- [x] Initialize Go module (`go mod init github.com/.../novel-tui`) and install dependencies (`github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`, `github.com/charmbracelet/bubbles`). <!-- sdd-owner: implementation -->
- [x] Implement domain models and repository interfaces in `internal/domain/domain.go`. <!-- sdd-owner: implementation -->
- [x] Implement filesystem chapter repository (`FileChapterRepository`) with atomic writes in `internal/repository/chapter_repo.go`. <!-- sdd-owner: implementation -->
- [x] Implement filesystem character repository (`FileCharacterRepository`) in `internal/repository/character_repo.go`. <!-- sdd-owner: implementation -->
- [x] Implement service layer for word count, reading time calculation, and business logic in `internal/service/metrics.go` and `internal/service/workspace.go`. <!-- sdd-owner: implementation -->
- [x] Implement UI theme, color palettes, and Lip Gloss styles in `internal/ui/theme/theme.go`. <!-- sdd-owner: implementation -->
- [x] Define typed Bubble Tea message types (`FocusMsg`, `ChapterSelectedMsg`, `TextChangedMsg`, `SaveRequestedMsg`, `SaveCompletedMsg`) in `internal/ui/messages/messages.go`. <!-- sdd-owner: implementation -->
- [x] Implement Sidebar component (`SidebarModel`) supporting chapter navigation, tab switching (Chapters/Lore), and character viewing in `internal/ui/components/sidebar.go`. <!-- sdd-owner: implementation -->
- [x] Implement Editor component (`EditorModel`) wrapping `bubbles/textarea` with line wrapping, live metrics emission, and save hotkeys in `internal/ui/components/editor.go`. <!-- sdd-owner: implementation -->
- [x] Implement Status Bar component (`StatusBarModel`) displaying title, metrics, reading time, and save indicators in `internal/ui/components/statusbar.go`. <!-- sdd-owner: implementation -->
- [x] Implement Root TUI model (`RootModel`) managing window sizing, minimum size warning check, focus cycling (`Tab`/`Shift-Tab`), and message routing in `internal/ui/model/root.go`. <!-- sdd-owner: implementation -->
- [x] Implement main entry point (`cmd/novel-tui/main.go`) to wire dependencies, initialize workspace folders, and run the Bubble Tea program. <!-- sdd-owner: implementation -->
- [x] Write unit tests for domain models, repositories, word-count/metric services, and UI state handlers in `internal/...` test files. <!-- sdd-owner: implementation -->
- [ ] Start or reuse bounded review. <!-- sdd-owner: parent -->
