# Implementation Tasks

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 1200 - 1800 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1: Domain, Config Repository & Workspace Manager -> PR 2: UI Messages, Reusable Modal & Launcher/LLM Components -> PR 3: Root Model State Machine refactor, Main AltScreen bootstrap & Tests |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

---

## Task Breakdown

- [x] 1. Implement domain models in `internal/domain/config.go` (`AppConfig`, `LLMConfig`, `NovelMetadata`) with JSON tags and defaults. <!-- sdd-owner: implementation -->
- [x] 2. Implement domain interfaces for spellchecking in `internal/domain/spellcheck.go` (`Spellchecker`, `DictionaryManager`). <!-- sdd-owner: implementation -->
- [x] 3. Write unit tests for domain models and default initialization in `internal/domain/config_test.go`. <!-- sdd-owner: implementation -->
- [x] 4. Implement Configuration Repository in `internal/repository/config_repo.go` for `~/.config/novel-tui/config.json` loading and saving. <!-- sdd-owner: implementation -->
- [x] 5. Write unit tests for `FileConfigRepository` in `internal/repository/config_repo_test.go`. <!-- sdd-owner: implementation -->
- [x] 6. Implement Multi-project Workspace Manager in `internal/service/workspace_manager.go` (discovery in root dir, slugification, novel scaffolding with `capitulos/`, `personajes.json`, `notas.txt`, and sequential chapter numbering). <!-- sdd-owner: implementation -->
- [x] 7. Write unit tests for `WorkspaceManager` in `internal/service/workspace_manager_test.go`. <!-- sdd-owner: implementation -->
- [x] 8. Update Chapter Repository in `internal/repository/chapter_repo.go` to operate on `capitulos/*.txt` and `personajes.json`. <!-- sdd-owner: implementation -->
- [x] 9. Define typed Bubble Tea messages in `internal/ui/messages/messages.go` (`ChangeViewMsg`, `ShowModalMsg`, `HideModalMsg`, `CreateNovelMsg`, `CreateChapterMsg`, `SaveLLMConfigMsg`, `OpenNovelMsg`, etc.). <!-- sdd-owner: implementation -->
- [x] 10. Implement reusable Modal Input Dialog component in `internal/ui/components/modal.go` with text input and validation. <!-- sdd-owner: implementation -->
- [x] 11. Implement Launcher Dashboard component in `internal/ui/components/launcher.go` with ASCII banner, recent novels list, and single-key action menu. <!-- sdd-owner: implementation -->
- [x] 12. Implement LLM Settings component in `internal/ui/components/llm_config.go` with interactive text input form for Ollama parameters and genre prompts. <!-- sdd-owner: implementation -->
- [x] 13. Refactor Root Model in `internal/ui/model/root.go` with View State Machine (`ViewStateLauncher`, `ViewStateEditor`, `ViewStateLLMConfig`, `ActiveModal`), transition commands, and key routing. <!-- sdd-owner: implementation -->
- [x] 14. Update `cmd/novel-tui/main.go` to load global configuration, initialize services, and start Bubble Tea with `tea.WithAltScreen()`. <!-- sdd-owner: implementation -->
- [x] 15. Write unit and component tests for all new components, services, and UI state transitions. <!-- sdd-owner: implementation -->
