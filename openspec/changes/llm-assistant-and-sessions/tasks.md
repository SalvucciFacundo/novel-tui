# Tasks: LLM Assistant & Persistent Sessions

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 1200 - 1800 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1: Domain, Repository & Providers (Core Backend) → PR 2: Context Builder & Session Service → PR 3: UI Components, Messages & Root Model Integration |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

---

## Task List

### Phase 1: Core Domain & Repository Layer
- [x] Implement domain models and interfaces in `internal/domain/llm.go` (`LLMProvider`, `ChatMessage`, `ChatRequest`, `StreamChunk`, `LLMEffortLevel`). <!-- sdd-owner: implementation -->
- [x] Implement domain session entity in `internal/domain/session.go` (`ChatSession`). <!-- sdd-owner: implementation -->
- [x] Implement JSON file-based session repository in `internal/repository/session_repo.go` (automatic `<novel_dir>/chats/` creation, `Save`, `Get`, `List`, `Delete`). <!-- sdd-owner: implementation -->
- [x] Add unit tests for session repository in `internal/repository/session_repo_test.go`. <!-- sdd-owner: implementation -->

### Phase 2: LLM Providers & Context Builder Service
- [x] Implement Ollama streaming client in `internal/service/llm/ollama.go` with NDJSON streaming and context cancellation. <!-- sdd-owner: implementation -->
- [x] Implement OpenAI-compatible streaming client in `internal/service/llm/openai.go` with SSE streaming and bearer auth. <!-- sdd-owner: implementation -->
- [x] Implement provider factory in `internal/service/llm/factory.go`. <!-- sdd-owner: implementation -->
- [x] Implement context compilation builder in `internal/service/llm/context_builder.go` (aggregates genre, `personajes.json`, `notas.txt`, active chapter, effort level instructions, and token budget truncation). <!-- sdd-owner: implementation -->
- [x] Implement comprehensive unit tests with mock HTTP test servers for streaming providers in `internal/service/llm/llm_test.go`. <!-- sdd-owner: implementation -->

### Phase 3: UI Messages & Chat Drawer Component
- [x] Define custom TUI messages in `internal/ui/messages/messages.go` (`ToggleChatDrawerMsg`, `SendChatMessageMsg`, `TokenReceivedMsg`, `StreamFinishedMsg`, `StreamErrorMsg`, `SelectSessionMsg`, `CreateSessionMsg`, `SetEffortLevelMsg`). <!-- sdd-owner: implementation -->
- [x] Implement Bubble Tea Chat Drawer component in `internal/ui/components/chat_drawer.go` (viewport history, textarea input, effort selector badge, session manager picker, non-blocking token chunk accumulator). <!-- sdd-owner: implementation -->

### Phase 4: Root Model Integration & Verification
- [x] Integrate Chat Drawer and 3-panel split layout (Sidebar + Editor + ChatDrawer) in `internal/ui/model/root.go` supporting `Ctrl+A` toggle, focus management (`FocusChat`), and window resizing. <!-- sdd-owner: implementation -->
- [x] Add component unit tests in `internal/ui/components/components_test.go` and root model tests in `internal/ui/model/root_test.go`. <!-- sdd-owner: implementation -->
- [x] Verify complete integration with `go test ./...`. <!-- sdd-owner: implementation -->
