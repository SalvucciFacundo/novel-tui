# Proposal: LLM Assistant & Persistent Sessions

## 1. Intent
The goal of this change is to introduce a fully integrated, non-blocking AI writing assistant directly into the novel-tui editor. By providing a multi-provider backend (supporting local models via Ollama and remote APIs via OpenAI-compatible endpoints) and a dedicated chat drawer, authors will be able to brainstorm, co-write, and evaluate their prose without leaving the terminal or losing context. Persistent chat sessions guarantee that narrative discussions are saved alongside the novel.

## 2. Scope
**Included in this change:**
- **Unified Multi-Provider Architecture**: An `LLMProvider` interface in `internal/domain` defining a `StreamChat` method using Go channels/context. Implementations for `OllamaProvider` (native API) and `OpenAICompatibleProvider`.
- **Configurable Effort Levels**: Low (quick tips), Medium (co-writing), and High (deep reasoning) settings that adjust system prompts and/or API parameters.
- **Persistent Chat Sessions**: A service and repository layer to save/load chat histories as JSON inside `<novel_dir>/chats/<session_id>.json`.
- **Genre-Adaptive Prompting & Context Injection**: A context builder that dynamically retrieves `personajes.json`, active chapter text, and `notas.txt`, wrapping them with the configured genre prompt.
- **TUI Chat Drawer Component**: A right-side drawer inside the Editor view, toggled via `Ctrl+A`. Supports streaming tokens via Bubble Tea commands (`tea.Cmd`) without blocking the main event loop, and allows switching effort levels and sessions.

**Non-goals (Out of scope):**
- Complex RAG (Retrieval-Augmented Generation) with vector databases; context injection will rely on raw text concatenation and simple truncation for now.
- Autonomous file editing (the LLM will provide chat responses; the user must manually copy/apply suggestions to the editor).

## 3. Affected Areas
- **`internal/domain`**: New models for `ChatMessage`, `ChatSession`, `LLMEffortLevel`, and the `LLMProvider` interface.
- **`internal/repository`**: New `ChatSessionRepository` for JSON file persistence.
- **`internal/service`**: New `LLMService` to coordinate the provider, context injection, and session management.
- **`internal/ui/components`**: New `ChatDrawerModel` and modifications to `EditorModel`/`RootModel` to manage the split-pane layout and `Ctrl+A` keybinding.
- **`internal/ui/messages`**: New message types for streaming tokens (`TokenReceivedMsg`, `StreamFinishedMsg`, `StreamErrorMsg`).

## 4. Risks & Tradeoffs
- **Event Loop Blocking**: If the TUI waits synchronously for LLM API responses, the terminal will freeze. The solution is strictly using asynchronous Go channels and Bubble Tea `tea.Cmd` message passing.
- **Terminal Real Estate**: A side-by-side drawer will consume significant horizontal space. On small terminals (under 100 columns), the editor area may become uncomfortably narrow.
- **Context Window Exhaustion**: Simply appending all chapters and characters into the prompt could easily exceed context limits (e.g., 4k or 8k tokens) and result in API errors or degraded responses.
- **Provider Differences**: Handling SSE (Server-Sent Events) streams for OpenAI-compatible APIs vs. Ollama's NDJSON streams requires careful parsing to avoid token delays or rendering glitches.

## 5. Rollback Plan
Since the chat sessions are cleanly scoped to a new subdirectory (`chats/`) within the workspace, rolling back involves:
1. Reverting the UI changes (removing `Ctrl+A` bindings and the drawer view from the editor).
2. Removing the new `internal/service/llm.go` and provider implementations.
3. Chat histories (`chats/*.json`) will remain harmlessly on disk without affecting normal novel operations.

## 6. Success Criteria
- The user can press `Ctrl+A` in the editor to open a right-side chat drawer.
- The user can select "Low", "Medium", or "High" effort levels in the drawer.
- Sending a message streams the response back token-by-token without stuttering or freezing cursor movement in the editor.
- The assistant accurately references details from `personajes.json` or the currently active chapter.
- Quitting the application and reopening it restores previous chat sessions from `<novel_dir>/chats/`.

---

## Proposal question round

To refine this PRD and ensure we build exactly what the workflow needs, please review these product questions and assumptions:

1. **Context limits & Truncation**: When injecting context (`capitulos/*.txt`), novels can be massive. Should we inject only the *active* chapter + `personajes.json`, or attempt to inject all chapters until we hit a hard character limit?
2. **Effort & Reasoning Levels**: Should the effort level (Low, Medium, High) control only the prompt instructions, or should it also modify API parameters like temperature and `max_tokens` (e.g., lower temp for High/Reasoning, higher temp for Medium/Creative)?
3. **UI Layout constraints**: The Right-hand drawer will reduce the Editor width. On smaller terminal windows, should the drawer permanently split the pane, or float over the text?
4. **Session switching**: To keep the initial slice lean, should the UI provide a simple cycling command (`Tab` through sessions in the drawer) or do we need a full list/menu for chat histories?

*(Please answer, skip, or correct any of these assumptions so we can finalize the proposal.)*