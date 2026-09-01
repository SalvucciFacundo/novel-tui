# Architecture & Design: novel-tui-layout

## 1. Architecture Overview

`novel-tui` follows the **Go Standard Project Layout** mixed with **The Elm Architecture** provided by the `charmbracelet/bubbletea` framework. The system is cleanly layered to separate UI components, domain logic, and file persistence.

### Directory Structure

```text
novel-tui/
├── cmd/
│   └── novel-tui/           # Main entry point, wire dependencies, start bubbletea program
├── internal/
│   ├── domain/              # Core domain models and interfaces
│   ├── repository/          # Data access layer (File System implementations)
│   ├── service/             # Business logic (e.g., metric calculations)
│   └── ui/                  # Bubble Tea frontend components
│       ├── model/           # Root TUI state and main update loop
│       ├── components/      # Sub-models (Sidebar, Editor, StatusBar)
│       ├── messages/        # Typed tea.Msg definitions
│       └── theme/           # Lip Gloss styling and layout definitions
└── data/                    # (Workspace) User's local workspace
```

## 2. Domain Models (`internal/domain`)

The domain layer encapsulates the core entities of a novel project without any UI coupling.

```go
package domain

// Chapter represents a single writing unit.
type Chapter struct {
    ID        string
    Title     string
    FilePath  string // Relative to project root
    WordCount int
    Content   string // Loaded on demand when selected
}

// Character tracks lore and context for the secondary sidebar tab.
type Character struct {
    ID          string
    Name        string
    Role        string
    Description string
    Notes       string
}

// EditorMetrics represents real-time typing metrics.
type EditorMetrics struct {
    WordCount   int
    CharCount   int
    ReadingTime int // in minutes (based on 200-250 wpm)
    IsDirty     bool
}

// Repositories define data persistence contracts.
type ChapterRepository interface {
    ListAll() ([]Chapter, error)
    LoadContent(id string) (string, error)
    SaveContent(id string, content string) error
    Create(title string) (Chapter, error)
}

type CharacterRepository interface {
    ListAll() ([]Character, error)
    SaveAll(chars []Character) error
}
```

## 3. UI Component Architecture (`internal/ui`)

The application uses a strict Parent-Child composition pattern (`RootModel` -> `SubModels`).

### `RootModel` (`internal/ui/model/root.go`)
- **State**: Tracks `width`, `height`, and `activeFocus` (enum: `FocusSidebar`, `FocusEditor`).
- **Update**: Intercepts global keys (`Ctrl+C` to quit, `Tab`/`Shift+Tab` to cycle focus). Responds to `tea.WindowSizeMsg` by calculating layout dimensions and passing size subsets down to child components. 
- **View**: Uses Lip Gloss flexbox-style joining to stitch the Sidebar, Editor, and Status Bar together.

### `SidebarModel` (`internal/ui/components/sidebar.go`)
- **State**: Tracks `activeTab` (`TabChapters` vs `TabLore`), lists of `[]domain.Chapter` and `[]domain.Character`, and internal list cursors.
- **Update**: Handles `Up/Down` navigation when focused. Handles `[` / `]` for tab switching. Emits `messages.ChapterSelectedMsg` when `Enter` is pressed on a chapter.

### `EditorModel` (`internal/ui/components/editor.go`)
- **Wrapper**: Wraps `charmbracelet/bubbles/textarea`.
- **State**: Holds the currently loaded chapter ID and the active `textarea.Model`.
- **Update**: When receiving typed input, updates the internal `textarea`. Emits `messages.TextChangedMsg` so the Status Bar can update. Handles `Ctrl+S` by emitting a `messages.SaveRequestedMsg`.

### `StatusBarModel` (`internal/ui/components/statusbar.go`)
- **State**: Caches current `domain.EditorMetrics` and current chapter title.
- **View**: Renders left-aligned info (Chapter name, Save Status) and right-aligned metrics (Words, Chars, Read Time) styled via Lip Gloss.

## 4. Message Flow (`internal/ui/messages`)

Communication between decoupled UI components happens entirely via typed `tea.Msg` structs passed through the central update loop.

- `FocusMsg{ Target: FocusState }`: Sent by `RootModel` to visually highlight/unhighlight components.
- `ChapterSelectedMsg{ Chapter: domain.Chapter }`: Emitted by `Sidebar`. Caught by `Editor` to load text into the buffer, and by `StatusBar` to update the title.
- `TextChangedMsg{ Metrics: domain.EditorMetrics }`: Emitted by `Editor` upon buffer modifications. Caught by `StatusBar` to display live metric updates and flip the save indicator to `[Modified*]`.
- `SaveRequestedMsg{ ChapterID, Content }`: Emitted by `Editor` on `Ctrl+S`. Caught by `RootModel` to invoke the `ChapterRepository.SaveContent` method via a `tea.Cmd`.
- `SaveCompletedMsg{ Success: bool, Error }`: Emitted after persistence finishes. Updates the status bar to `[Saved]`.

## 5. Persistence Layer (`internal/repository`)

The persistence logic executes asynchronously as Bubble Tea commands (`tea.Cmd`) to avoid blocking the main TUI render thread.

- **`FileChapterRepository`**: 
  - Writes to `<project_root>/chapters/<slug>.md`.
  - Implements atomic writes (writing to a `.tmp` file and renaming) to prevent corruption if the terminal crashes mid-save.
- **`FileCharacterRepository`**:
  - Reads and writes to `<project_root>/characters.json`.
  - Serializes standard Go structs.

## 6. Theming System (`internal/ui/theme`)

Lip Gloss handles all colors, borders, and margins.

- **Palette**: Define a common struct (e.g., `theme.Catppuccin` or `theme.Nord`) for foregrounds, backgrounds, accents, and muted text.
- **Borders**: 
  - `FocusedBorder()`: Uses active accent color, normal borders.
  - `BlurredBorder()`: Uses muted colors, dim borders.
- **Badges**: Status bar elements use inverted pills (e.g., `lipgloss.NewStyle().Background(Accent).Foreground(Base).Render(" SAVED ")`).

## 7. Edge Cases & Constraints
- **Too Small Terminal**: If `WindowSizeMsg.Width < 60` or `Height < 15`, `RootModel.View()` skips rendering the normal layout and instead renders a centered text message: "Please resize your terminal (Min: 60x15)."
- **Concurrency**: Text metric recalculation runs on every keystroke. For massive files, if this causes input lag, metric recalculation must be decoupled into a debounced `tea.Cmd`. (Deferred optimization, standard `textarea` handles moderate text fine).
