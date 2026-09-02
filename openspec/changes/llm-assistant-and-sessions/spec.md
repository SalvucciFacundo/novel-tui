# LLM Assistant & Persistent Sessions Specification

## Purpose

Define the functional requirements, domain interfaces, persistence rules, prompt and context compilation mechanisms, and TUI component behaviors for the integrated AI Writing Assistant in `novel-tui`. This enables authors to co-write, brainstorm, and critique prose with local or remote LLMs directly from the terminal editor without blocking UI responsiveness or losing narrative discussion histories.

---

## Requirements

### Requirement: Multi-Provider Streaming Contract

The system MUST define a unified `LLMProvider` domain interface and provide concrete client implementations for Ollama (`OllamaProvider`) and OpenAI-compatible endpoints (`OpenAICompatibleProvider`), streaming chat response chunks asynchronously through Go channels with full `context.Context` cancellation support.

#### Scenario: Unified domain interface contract
- GIVEN the domain package `internal/domain`
- WHEN the LLM provider contract is declared
- THEN it MUST define:
  ```go
  type LLMProvider interface {
      StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
  }
  ```
- AND `ChatRequest` MUST contain:
  - `Messages`: slice of `ChatMessage` (system, user, assistant)
  - `Model`: string identifier for the target model
  - `Temperature`: float64 generation temperature
  - `MaxTokens`: integer token limit (optional/nullable)
- AND `StreamChunk` MUST contain:
  - `Content`: string delta token received from the backend
  - `Done`: boolean flag indicating the end of the stream
  - `Err`: error representing network, decoding, or provider failures

#### Scenario: Ollama native streaming
- GIVEN an active `OllamaProvider` configured with `BaseURL` (e.g., `http://localhost:11434`) and model `qwen2.5:7b`
- WHEN `StreamChat` is invoked with a valid `ChatRequest`
- THEN the provider MUST issue an HTTP `POST` request to `{BaseURL}/api/chat` with JSON body `{ "model": "...", "messages": [...], "stream": true, "options": { "temperature": ... } }`
- AND parse the incoming NDJSON stream line by line
- AND deliver each `message.content` fragment into the returned `StreamChunk` channel
- AND close the channel after the final chunk where `done: true` is received.

#### Scenario: OpenAI-compatible SSE streaming
- GIVEN an active `OpenAICompatibleProvider` configured with `BaseURL` (e.g., `https://api.openai.com/v1`, OpenRouter, or local vLLM endpoint) and optional `APIKey`
- WHEN `StreamChat` is invoked with a valid `ChatRequest`
- THEN the provider MUST issue an HTTP `POST` request to `{BaseURL}/chat/completions` with header `Authorization: Bearer <APIKey>` (when key is present) and body `{ "model": "...", "messages": [...], "stream": true, "temperature": ... }`
- AND parse the incoming `text/event-stream` (SSE) data lines starting with `data: `
- AND extract delta tokens from `choices[0].delta.content`
- AND stop streaming and close the channel when receiving the `data: [DONE]` sentinel line.

#### Scenario: Stream cancellation via context
- GIVEN an ongoing LLM generation streaming tokens via channel
- WHEN the caller cancels the associated `context.Context` (e.g., user hits cancel or closes the drawer)
- THEN the provider MUST abort the underlying HTTP request immediately
- AND drain/close the Go channel without leaking goroutines or network connections.

#### Scenario: Provider connection error handling
- GIVEN an unreachable endpoint or invalid authorization token
- WHEN `StreamChat` attempts to connect or receives an HTTP error status (e.g., 401, 404, 500)
- THEN `StreamChat` MUST return an error or emit a `StreamChunk` with non-nil `Err` containing a human-readable diagnostic message
- AND close the channel cleanly without crashing the application.

---

### Requirement: Configurable Effort and Reasoning Levels

The system MUST support three distinct LLM Effort Levels (`Low`, `Medium`, `High`) that dynamically modify system instructions, generation parameters, and reasoning depth per chat session or interaction.

#### Scenario: Effort Level definitions and parameters
- GIVEN the domain enum `LLMEffortLevel`
- WHEN effort levels are configured
- THEN the system MUST support the following profiles:
  - `EffortLow` ("Bajo"): Focused on concise suggestions, quick copy edits, and rapid micro-feedback. Sets lower temperature (e.g., 0.3–0.5), restrictive length guidance, and prompt instruction: `"Sé conciso, directo y breve. Responde en 1-3 frases máximo."`
  - `EffortMedium` ("Medio"): Standard co-writing partner. Sets balanced temperature (e.g., 0.7), standard creative length, and prompt instruction: `"Actúa como co-escritor creativo. Sugiere mejoras estilísticas y extensiones de prosa manteniendo la voz del autor."`
  - `EffortHigh` ("Alto / Razonamiento"): Deep literary analysis and critique. Sets analytical temperature (e.g., 0.4–0.6), chain-of-thought analysis directive, and prompt instruction: `"Actúa como editor literario senior. Analiza estructura narrativa, arcos de personajes, subtramas y coherencia causal paso a paso antes de dar recomendaciones detalladas."`

#### Scenario: Dynamic effort switching in active session
- GIVEN an open chat session with current effort level set to `EffortMedium`
- WHEN the user toggles or selects `EffortHigh` via the TUI effort selector
- THEN the session's active effort level MUST update immediately to `EffortHigh`
- AND subsequent user queries in this session MUST be dispatched using the `EffortHigh` prompt instructions and parameter profile
- AND the updated effort level MUST be persisted to the session file.

---

### Requirement: Persistent Chat Sessions

The system MUST persist full conversation histories and session metadata inside the novel project under `<novel_dir>/chats/<session_id>.json`, allowing authors to revisit, continue, switch, and delete chat threads across editing sessions.

#### Scenario: Storage layout and automatic directory initialization
- GIVEN a loaded novel project located at `<novel_dir>`
- WHEN the chat session repository is accessed or initialized
- THEN the system MUST ensure the directory `<novel_dir>/chats/` exists, creating it automatically if missing
- AND store each session as an independent JSON file named `<session_id>.json` (where `<session_id>` is a unique slug or UUID).

#### Scenario: Chat session data schema
- GIVEN a chat session being serialized to `<novel_dir>/chats/<session_id>.json`
- WHEN the repository saves the file
- THEN the JSON content MUST strictly follow the structure:
  ```json
  {
    "id": "session_20250510_153000_plot_twist",
    "title": "Discusión sobre el giro argumental del Cap 2",
    "novel_path": "/path/to/novel",
    "effort_level": "medium",
    "created_at": "2025-05-10T15:30:00Z",
    "updated_at": "2025-05-10T15:35:12Z",
    "messages": [
      {
        "role": "user",
        "content": "¿Cómo puedo justificar la traición de Mateo?",
        "timestamp": "2025-05-10T15:30:00Z"
      },
      {
        "role": "assistant",
        "content": "Podrías sembrar pistas sutiles en el capítulo 1...",
        "timestamp": "2025-05-10T15:30:04Z"
      }
    ]
  }
  ```

#### Scenario: Automatic title generation on first message
- GIVEN a new, untitled chat session
- WHEN the user sends the first prompt message (e.g., `"Revisar diálogo entre Elena y Kuno"`)
- THEN the system MUST derive an initial title from the prompt (truncated to 40 characters)
- AND persist the updated session metadata to `<novel_dir>/chats/<session_id>.json`.

#### Scenario: Listing and loading existing sessions
- GIVEN `<novel_dir>/chats/` contains multiple valid `.json` session files
- WHEN the chat repository `List` method is executed
- THEN it MUST return all sessions ordered by `updated_at` descending (most recent first)
- AND loading a specific session by ID MUST deserialize all historical messages into memory ready for display and conversation resumption.

#### Scenario: Deleting a chat session
- GIVEN an existing chat session file `<novel_dir>/chats/<session_id>.json`
- WHEN the user confirms deletion of this session
- THEN the repository MUST remove the corresponding `.json` file from disk
- AND update the active UI state to load the next most recent session or create an empty new session if none remain.

---

### Requirement: Genre-Adaptive System Prompts & Context Builder

The system MUST provide a `ContextBuilder` that aggregates genre-specific templates, project lore from `personajes.json`, author notes from `notas.txt`, and active chapter text from `capitulos/*.txt`, compiling them into a structured context payload before sending queries to the LLM.

#### Scenario: Compiling full narrative context
- GIVEN a novel with:
  - Configured genre: `"Fantasía Oscura"`
  - Character file `<novel_dir>/personajes.json` containing characters `Elena` and `Mateo`
  - Notes file `<novel_dir>/notas.txt` containing plot guidelines
  - Active chapter `<novel_dir>/capitulos/02_la_emboscada.txt` currently open in editor
- WHEN the user submits a query to the AI assistant
- THEN the `ContextBuilder` MUST assemble a composite system prompt containing:
  1. Base Persona & Genre Guidelines (e.g., dark fantasy worldbuilding, internal consistency, tone).
  2. Selected Effort Level directive (`Low`, `Medium`, or `High`).
  3. Character Lore Block: serialized character names, roles, descriptions, and relationships from `personajes.json`.
  4. Author Notes Block: text extracted from `notas.txt`.
  5. Active Chapter Context: current text of `02_la_emboscada.txt` wrapped in explicit context markers (e.g., `--- INICIO CAPÍTULO ACTUAL (02_la_emboscada.txt) ---`).

#### Scenario: Handling missing or empty project files
- GIVEN a novel workspace where `personajes.json` is empty `[]`, `notas.txt` is empty, or no chapter is currently loaded
- WHEN the `ContextBuilder` generates the context payload
- THEN the system MUST NOT crash or fail
- AND omit the empty sections cleanly while retaining the genre prompt and effort level instructions.

#### Scenario: Context truncation and budget safety
- GIVEN an active chapter or note file whose combined length exceeds the safe context threshold (e.g., > 12,000 characters / ~3,000 tokens)
- WHEN the `ContextBuilder` compiles the context
- THEN the system MUST truncate older chapter text or provide a focused window around the cursor/recent paragraphs
- AND append a notice `[...contenido anterior truncado por límites de contexto...]` to preserve LLM token window integrity.

---

### Requirement: Interactive Chat Drawer in Editor View

The system MUST render a dedicated split-pane Chat Drawer on the right side of the Editor view, toggled via `Ctrl+A`, supporting non-blocking token streaming via Bubble Tea commands (`tea.Cmd`), session switching, and effort level adjustment.

#### Scenario: Toggling Chat Drawer visibility
- GIVEN the user is in the Editor view with cursor in the main writing area
- WHEN the user presses `Ctrl+A`
- THEN the system MUST toggle the visibility of the right-side Chat Drawer
- AND when opened:
  - Split the screen horizontally into Editor (left) and Chat Drawer (right, taking e.g., 35-40% width or min 35 columns)
  - Transfer keyboard focus to the Chat Drawer text input
- AND when closed:
  - Restore the Editor to full width
  - Return keyboard focus to the main text editor at the previous cursor position.

#### Scenario: Non-blocking token streaming in Bubble Tea
- GIVEN the Chat Drawer is open and the user submits a message with `Enter`
- WHEN the LLM begins generating tokens
- THEN the UI MUST dispatch a non-blocking `tea.Cmd` reading from the Go channel
- AND emit `TokenReceivedMsg` for each arriving token delta
- AND append incoming tokens to the assistant message in real-time
- AND keep the Bubble Tea event loop responsive (allowing scrolling, typing, or status bar updates without UI freezing)
- AND dispatch `StreamFinishedMsg` when generation concludes.

#### Scenario: Cancelling an active generation in drawer
- GIVEN an ongoing streaming response in the Chat Drawer
- WHEN the user presses `Esc` while generation is active
- THEN the system MUST cancel the stream context
- AND finalize the assistant message with the tokens received up to that point
- AND return the drawer to the ready input state.

#### Scenario: Effort Level Selector in Drawer
- GIVEN the Chat Drawer is focused
- WHEN the user presses the effort cycle shortcut (e.g., `Ctrl+E` or `Tab` when focused on effort badge)
- THEN the effort level MUST cycle between `[Bajo]`, `[Medio]`, and `[Alto]`
- AND visually update the drawer header badge immediately
- AND apply the new effort level to subsequent messages.

#### Scenario: Session switching and new session creation in Drawer
- GIVEN the Chat Drawer is open
- WHEN the user presses `Ctrl+S` (or triggers the session menu shortcut within the drawer)
- THEN the drawer MUST present a quick-picker list of saved sessions for this novel
- AND selecting a session MUST load its full message history into the viewport
- AND pressing `Ctrl+N` within the drawer MUST create a blank new session and clear the message viewport for a fresh conversation.

#### Scenario: Responsive terminal resizing with open Drawer
- GIVEN the Chat Drawer is open
- WHEN the terminal window is resized (`tea.WindowSizeMsg`)
- THEN if width >= 90 columns, the system MUST recalculate dimensions to maintain split proportion (e.g. Editor ~60%, Drawer ~40%)
- AND if width < 90 columns, the system MUST prioritize the focused pane or render a compact floating/overlay drawer with minimum 30 columns width to prevent unreadable word wrapping.

---
