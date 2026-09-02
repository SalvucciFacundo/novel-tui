# Apply Progress: LLM Assistant & Persistent Sessions

## Status
- Implementation: Complete
- Verification: Passed (`go test -count=1 ./...` and `go build ./cmd/novel-tui`)

## Completed Tasks
- [x] Implement domain models and interfaces in `internal/domain/llm.go` (`LLMProvider`, `ChatMessage`, `ChatRequest`, `StreamChunk`, `LLMEffortLevel`).
- [x] Implement domain session entity in `internal/domain/session.go` (`ChatSession`, `ChatSessionRepository`).
- [x] Implement JSON file-based session repository in `internal/repository/session_repo.go` (auto `<novel_dir>/chats/` creation, `Save`, `Get`, `List`, `Delete`, auto-title generation).
- [x] Add unit tests for session repository in `internal/repository/session_repo_test.go`.
- [x] Implement Ollama streaming client in `internal/service/llm/ollama.go` with NDJSON streaming and context cancellation.
- [x] Implement OpenAI-compatible streaming client in `internal/service/llm/openai.go` with SSE streaming and bearer auth.
- [x] Implement provider factory in `internal/service/llm/factory.go`.
- [x] Implement context compilation builder in `internal/service/llm/context_builder.go` (aggregates genre, `personajes.json`, `notas.txt`, active chapter, effort level instructions, and token budget truncation).
- [x] Implement comprehensive unit tests with mock HTTP test servers for streaming providers in `internal/service/llm/llm_test.go`.
- [x] Define custom TUI messages in `internal/ui/messages/messages.go` (`ToggleChatDrawerMsg`, `SendChatMessageMsg`, `TokenReceivedMsg`, `StreamFinishedMsg`, `StreamErrorMsg`, `SelectSessionMsg`, `CreateSessionMsg`, `SetEffortLevelMsg`, `FocusChat`).
- [x] Implement Bubble Tea Chat Drawer component in `internal/ui/components/chat_drawer.go` (viewport history, textarea input, effort selector badge, session manager picker, non-blocking token chunk accumulator).
- [x] Integrate Chat Drawer and 3-panel split layout (Sidebar + Editor + ChatDrawer) in `internal/ui/model/root.go` supporting `Ctrl+A` toggle, focus management (`FocusChat`), and window resizing.
- [x] Add component unit tests in `internal/ui/components/components_test.go` and root model tests in `internal/ui/model/root_test.go`.
- [x] Verify complete integration with `go test ./...` and `go build ./cmd/novel-tui`.

## Files Changed
- `internal/domain/llm.go` (created)
- `internal/domain/session.go` (created)
- `internal/repository/session_repo.go` (created)
- `internal/repository/session_repo_test.go` (created)
- `internal/service/llm/ollama.go` (created)
- `internal/service/llm/openai.go` (created)
- `internal/service/llm/factory.go` (created)
- `internal/service/llm/context_builder.go` (created)
- `internal/service/llm/llm_test.go` (created)
- `internal/ui/messages/messages.go` (modified)
- `internal/ui/components/chat_drawer.go` (created)
- `internal/ui/components/components_test.go` (modified)
- `internal/ui/model/root.go` (modified)
- `internal/ui/model/root_test.go` (modified)
- `openspec/changes/llm-assistant-and-sessions/tasks.md` (modified)

## Test Commands Run
- `go test -v -count=1 ./...` -> All tests PASS
- `go build ./cmd/novel-tui` -> Succeeded

## Workload & PR Boundary
- Full change implemented and verified across domain, repository, service, and UI layers.
- Ready for verify phase.
