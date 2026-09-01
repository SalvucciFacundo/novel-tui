# Architecture Design: Launcher & Workspace Config

## 1. Executive Summary

This architecture design specifies the state machine refactoring, domain models, workspace services, and UI components required to transform `novel-tui` into a multi-project workspace manager. The design adheres strictly to Clean/Hexagonal Architecture principles, decoupling the BubbleTea UI from the workspace file operations and configuration management. It introduces a comprehensive state machine in the root BubbleTea model, supporting a Launcher dashboard, an LLM configuration screen, reusable input modals, and the foundational domain interfaces for spellchecking.

## 2. Hexagonal Architecture / Domain-Driven Design (DDD)

Logic is partitioned into Domain models (core business logic), Repositories (persistence layer), Services (orchestration and file-system workflows), and UI (terminal presentation).

### 2.1 Domain Layer

**`internal/domain/config.go`**
- **`AppConfig`**: The root configuration entity. Stores `RootDir` (string, defaults to `~/Novelas`), `RecentNovels` (slice of strings or paths), and an embedded `LLMConfig`.
- **`LLMConfig`**: Holds `Provider` (string), `BaseURL` (string), `Model` (string), `Temperature` (float64), and `GenrePrompts` (map[string]string).
- **`NovelMetadata`**: Encapsulates data discovered by the workspace manager: `Title`, `AbsolutePath`, `ChapterCount`, and `LastModified`.

**`internal/domain/spellcheck.go`**
Defines interfaces decoupled from the UI and CGO dependencies:
- **`Spellchecker` Interface**:
  - `Check(word string) bool`
  - `Suggestions(word string) []string`
- **`DictionaryManager` Interface**:
  - `LoadDictionary(affPath, dicPath string) error`
  - `AddCustomWord(word string) error`
  - `AvailableDictionaries() []string`

### 2.2 Repository Layer

**`internal/repository/config_repo.go`**
- **`ConfigRepository` Interface**: `Load() (*domain.AppConfig, error)`, `Save(cfg *domain.AppConfig) error`.
- **`FileConfigRepository`**: Concrete implementation utilizing `encoding/json` and `os` package to persist the config exactly to `~/.config/novel-tui/config.json`.

### 2.3 Service Layer

**`internal/service/workspace_manager.go`**
Responsible for multi-project scaffolding and discovery.
- **`WorkspaceManager` Interface**:
  - `Initialize(rootDir string) error`: Ensures the root directory exists.
  - `ListRecentNovels(rootDir string) ([]domain.NovelMetadata, error)`: Scans the root directory.
  - `CreateNovel(rootDir, title string) (*domain.NovelMetadata, error)`: Sanitizes the title (slugification), creates the folder `<root_dir>/<slug>/`, and scaffolds the standard structure (`capitulos/01_capitulo_1.txt`, `personajes.json`, `notas.txt`).
  - `CreateChapter(novelDir, chapterTitle string) (string, error)`: Resolves the next sequence number by reading `capitulos/`, formats the filename as `XX_<slug>.txt`, creates it, and returns the path.

## 3. UI Architecture (Bubble Tea)

The presentation layer utilizes the Elm-like architecture of `BubbleTea`.

### 3.1 View State Machine

**`internal/ui/model/root.go`**
The root model routes update messages and renders the active component.
- **States (Enum `ViewState`)**:
  - `ViewStateLauncher`: Displays the recent projects dashboard.
  - `ViewStateEditor`: The core text editor.
  - `ViewStateLLMConfig`: The LLM settings form.
- **Modal Overlay**: A secondary `ActiveModal` boolean or sub-model that, when active, intercepts keyboard events before they reach the underlying ViewState model.
- **Transition Logic**: On receiving a `messages.ChangeViewMsg`, the root model updates `ViewState` and triggers a full screen re-render via `tea.WindowSizeMsg`.

### 3.2 UI Components

- **Launcher (`internal/ui/components/launcher.go`)**:
  - Renders the ASCII banner and application title.
  - Manages a `bubbles/list` for selecting recent novels.
  - Captures action keys (`c`, `n`, `o`, `l`, `d`, `q`) and dispatches commands like `messages.ShowModalMsg` or `messages.ChangeViewMsg`.
  
- **LLM Config (`internal/ui/components/llm_config.go`)**:
  - A form comprised of multiple `bubbles/textinput` elements.
  - Validates constraints (e.g., Temperature bounds).
  - Emits `messages.SaveLLMConfigMsg` to trigger repository persistence and `messages.ChangeViewMsg` to return to the Launcher.

- **Modal (`internal/ui/components/modal.go`)**:
  - A reusable floating window centered on the screen holding a single `bubbles/textinput`.
  - Driven by a `Purpose` context (e.g., `PurposeNewNovel`, `PurposeNewChapter`, `PurposeSetRootDir`).
  - Upon submission (Enter), it dispatches contextual messages (e.g., `messages.CreateNovelMsg{Title: input}`) and closes itself.

### 3.3 Typed Messages

**`internal/ui/messages/messages.go`**
All components communicate via strong-typed `tea.Msg` definitions:
- `ConfigLoadedMsg{Config *domain.AppConfig}`
- `ChangeViewMsg{View ViewState}`
- `ShowModalMsg{Purpose string, Prompt string, InitialValue string}`
- `HideModalMsg{}`
- `CreateNovelMsg{Title string}`
- `CreateChapterMsg{Title string}`
- `OpenNovelMsg{Path string}`
- `SaveLLMConfigMsg{Config domain.LLMConfig}`

## 4. Bootstrapping & AltScreen

**`cmd/novel-tui/main.go`**
1. Initialize the `FileConfigRepository` and `WorkspaceManager`.
2. Load global configuration, expanding the `~` path.
3. If executed with a directory argument, initialize `ViewStateEditor`. Otherwise, default to `ViewStateLauncher`.
4. Start the application explicitly with `tea.WithAltScreen()`:
   ```go
   p := tea.NewProgram(rootModel, tea.WithAltScreen())
   ```
   This guarantees that view transitions and final application exit don't leave artifacts in the terminal scrollback buffer.

## 5. Security & Edge Cases

- **Path Traversal Mitigation**: The `WorkspaceManager` must rigorously sanitize inputs during `CreateNovel` and `CreateChapter` to prevent users from typing `../../` and escaping the root directory.
- **Empty / Invalid Inputs**: The `modal.go` component will reject empty string submissions, displaying an inline error and refusing to dispatch the creation message.
- **Directory Clashes**: If a novel with the derived slug already exists, the `WorkspaceManager` returns a specific collision error, which the Modal renders without wiping user input.
