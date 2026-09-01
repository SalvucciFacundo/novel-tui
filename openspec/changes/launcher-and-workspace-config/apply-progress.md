# Apply Progress: Launcher and Workspace Config

## Completed Implementation Tasks

- [x] 1. Implement domain models in `internal/domain/config.go` (`AppConfig`, `LLMConfig`, `NovelMetadata`) with JSON tags and defaults.
- [x] 2. Implement domain interfaces for spellchecking in `internal/domain/spellcheck.go` (`Spellchecker`, `DictionaryManager`).
- [x] 3. Write unit tests for domain models and default initialization in `internal/domain/config_test.go`.
- [x] 4. Implement Configuration Repository in `internal/repository/config_repo.go` for `~/.config/novel-tui/config.json` loading and saving.
- [x] 5. Write unit tests for `FileConfigRepository` in `internal/repository/config_repo_test.go`.
- [x] 6. Implement Multi-project Workspace Manager in `internal/service/workspace_manager.go` (discovery in root dir, slugification, novel scaffolding with `capitulos/`, `personajes.json`, `notas.txt`, and sequential chapter numbering).
- [x] 7. Write unit tests for `WorkspaceManager` in `internal/service/workspace_manager_test.go`.
- [x] 8. Update Chapter Repository in `internal/repository/chapter_repo.go` to operate on `capitulos/*.txt` and `personajes.json`.
- [x] 9. Define typed Bubble Tea messages in `internal/ui/messages/messages.go` (`ChangeViewMsg`, `ShowModalMsg`, `HideModalMsg`, `CreateNovelMsg`, `CreateChapterMsg`, `SaveLLMConfigMsg`, `OpenNovelMsg`, etc.).
- [x] 10. Implement reusable Modal Input Dialog component in `internal/ui/components/modal.go` with text input and validation.
- [x] 11. Implement Launcher Dashboard component in `internal/ui/components/launcher.go` with ASCII banner, recent novels list, and single-key action menu.
- [x] 12. Implement LLM Settings component in `internal/ui/components/llm_config.go` with interactive text input form for Ollama parameters and genre prompts.
- [x] 13. Refactor Root Model in `internal/ui/model/root.go` with View State Machine (`ViewStateLauncher`, `ViewStateEditor`, `ViewStateLLMConfig`, `ActiveModal`), transition commands, and key routing.
- [x] 14. Update `cmd/novel-tui/main.go` to load global configuration, initialize services, and start Bubble Tea with `tea.WithAltScreen()`.
- [x] 15. Write unit and component tests for all new components, services, and UI state transitions.

## Files Changed

- `cmd/novel-tui/main.go`
- `internal/domain/config.go`
- `internal/domain/config_test.go`
- `internal/domain/spellcheck.go`
- `internal/repository/chapter_repo.go`
- `internal/repository/chapter_repo_test.go`
- `internal/repository/character_repo.go`
- `internal/repository/config_repo.go`
- `internal/repository/config_repo_test.go`
- `internal/service/workspace_manager.go`
- `internal/service/workspace_manager_test.go`
- `internal/ui/components/components_test.go`
- `internal/ui/components/launcher.go`
- `internal/ui/components/llm_config.go`
- `internal/ui/components/modal.go`
- `internal/ui/components/sidebar.go`
- `internal/ui/messages/messages.go`
- `internal/ui/model/root.go`
- `internal/ui/model/root_test.go`
- `openspec/changes/launcher-and-workspace-config/tasks.md`

## Verification Evidence

- `go test -v -count=1 ./...` (All tests in domain, repository, service, components, and model passed)
- `go build -v ./cmd/novel-tui` (Binary builds cleanly without compilation errors)
