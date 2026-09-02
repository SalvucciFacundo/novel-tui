# Architecture Design: LLM Assistant & Persistent Sessions

## 1. Overview
This design introduces a non-blocking AI writing assistant to `novel-tui`. The architecture is split into a robust domain layer defining the LLM contracts and session entities, a service layer handling multi-provider streaming and dynamic context building, a repository layer for persisting chat sessions, and a responsive TUI integration using Bubble Tea for seamless token streaming and window layout management.

## 2. Domain Models & Interfaces (`internal/domain`)
The domain layer isolates business rules from specific API clients and storage mechanisms.

### `internal/domain/llm.go`
Defines the `LLMProvider` interface and the streaming structures.

```go
package domain

import "context"

type LLMEffortLevel string

const (
    EffortLow    LLMEffortLevel = "low"
    EffortMedium LLMEffortLevel = "medium"
    EffortHigh   LLMEffortLevel = "high"
)

type StreamChunk struct {
    Content string
    Done    bool
    Err     error
}

type ChatMessage struct {
    Role      string    `json:"role"`
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp,omitempty"`
}

type ChatRequest struct {
    Messages    []ChatMessage
    Model       string
    Temperature float64
    MaxTokens   int
}

type LLMProvider interface {
    StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
}
```

### `internal/domain/session.go`
Defines the chat session entity.

```go
package domain

import "time"

type ChatSession struct {
    ID          string         `json:"id"`
    Title       string         `json:"title"`
    NovelPath   string         `json:"novel_path"`
    EffortLevel LLMEffortLevel `json:"effort_level"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    Messages    []ChatMessage  `json:"messages"`
}
```

## 3. Service & Providers (`internal/service/llm`)

### `internal/service/llm/factory.go`
A factory to initialize the correct provider (Ollama or OpenAI-compatible) based on the workspace configuration.

### `internal/service/llm/ollama.go`
Implements `domain.LLMProvider` for Ollama. Uses Go's `net/http` to send a POST request to `<base_url>/api/chat` with `stream: true`. Reads the NDJSON response line-by-line using `bufio.Scanner` and emits chunks to the returned Go channel. Supports early termination if the `context.Context` is cancelled.

### `internal/service/llm/openai.go`
Implements `domain.LLMProvider` for OpenAI-compatible endpoints (OpenAI, OpenRouter, vLLM). Sends a POST to `<base_url>/chat/completions` with SSE streaming enabled. Parses lines starting with `data: ` and unmarshals the JSON to extract `choices[0].delta.content`. Ends iteration when `data: [DONE]` is found.

### `internal/service/llm/context_builder.go`
Responsible for aggregating workspace context.
- Exposes `BuildContext(novelDir string, activeChapterText string, effort domain.LLMEffortLevel) string`.
- Reads `personajes.json` (character lore) and `notas.txt` (author notes).
- Applies truncation (e.g., limits context size to a fixed number of runes/characters) and constructs the final system prompt by combining these files with effort-specific instructions.

### `internal/service/llm/service.go`
Coordinates between the repository (saving/loading sessions), the context builder (preparing prompts), and the provider (streaming responses). 

## 4. Session Persistence (`internal/repository/session_repo.go`)
Implements persistence for `domain.ChatSession`.
- Uses `os.MkdirAll` to ensure `<novel_dir>/chats/` exists.
- Provides methods: `Save(session domain.ChatSession) error`, `List(novelDir string) ([]domain.ChatSession, error)`, `Get(novelDir, id string) (domain.ChatSession, error)`, and `Delete(novelDir, id string) error`.
- Writes JSON files named `<session_id>.json`.
- Titles are automatically generated (truncated to ~40 characters) from the first user message when `Title` is empty.

## 5. UI Layout & Event Routing (`internal/ui/model/root.go`)
The main layout of the TUI is expanding to handle a 3-pane structure.

- **Toggle Logic**: `root.go` captures `Ctrl+A`. When pressed, it toggles `model.showChatDrawer` boolean.
- **Window Sizing**: On `tea.WindowSizeMsg`, if `showChatDrawer` is true and width is >= 90 columns, the width is split (e.g., Sidebar: 20%, Editor: 45%, Drawer: 35%). If width < 90, the Drawer acts as a floating overlay or takes the bulk of the screen.
- **Event Delegation**: Keypresses are routed to the Chat Drawer when it is focused. 

## 6. Chat Drawer Component (`internal/ui/components/chat_drawer.go`)
A new Bubble Tea component (`ChatDrawerModel`) wrapping `viewport` (for message history display) and `textarea` (for user input).

- **State variables**: `session`, `isGenerating` (bool), `streamingContext` (context.Context), `cancelStream` (context.CancelFunc).
- **Keybindings**: 
  - `Enter`: Submits message to the LLM (if textarea isn't just inserting a newline).
  - `Esc`: Cancels active stream if `isGenerating` is true.
  - `Ctrl+E`: Cycles effort levels visually (Bajo -> Medio -> Alto).
  - `Ctrl+S` / `Ctrl+N`: Session picker and new session logic.
- **Update loop (`Update`)**: Dispatches `tea.Cmd` to the LLM service upon submit. Handles stream tokens recursively.

## 7. Non-Blocking Event Loop (Bubble Tea Messages)
To keep the TUI responsive, interaction with the LLM Service occurs through asynchronous Go channels managed by a long-running Bubble Tea command.

### `internal/ui/messages/messages.go`
Defines new message types:
```go
type StreamTokenMsg struct {
    Content string
}

type StreamFinishedMsg struct{}

type StreamErrorMsg struct {
    Err error
}
```

### Command Flow:
1. User presses `Enter`. Drawer adds User message to UI, sets `isGenerating = true`.
2. Drawer fires a `tea.Cmd` that calls `llmService.StreamChat(ctx)`.
3. Inside this command, a goroutine listens to the returned `<-chan domain.StreamChunk`.
4. For each chunk received, the command yields a `StreamTokenMsg` back to Bubble Tea using `tea.Batch` or by sending it to the main program loop. (A common pattern is a recursive command or `tea.Sequence` style reader).
5. When `Done: true`, it emits `StreamFinishedMsg`.
6. Drawer's `Update` catches `StreamTokenMsg`, appends the string to the active assistant message, and scrolls the viewport down.

## 8. Rollout & Migrations
- Safe migration: This change touches only new subsystems and the root view's resize layout. Existing novel files (`.txt`, `.json`) are read-only.
- New dependencies: The LLM services use standard library `net/http` and `bufio`; no massive heavy external SDKs (like `go-openai`) are strictly required if we consume the REST APIs directly, keeping binary size down.
