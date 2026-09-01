# Launcher and Workspace Configuration Specification

## Purpose

Define the functional and interface requirements for the Home Dashboard (Launcher), global configuration management, multi-project workspace scaffolding, LLM settings configuration, in-editor modal dialogs (such as chapter creation), and the core domain interfaces for spellchecking in `novel-tui`.

## Requirements

### Requirement: Launcher Dashboard View and Navigation

The system MUST render an interactive Home Dashboard when launched without an explicit project path, displaying an ASCII banner, a list of recent novels with metadata, and an action menu with single-key shortcuts.

#### Scenario: Initial launch without directory argument
- GIVEN the application is started without command-line directory arguments
- WHEN the root UI model initializes
- THEN the system MUST render the Launcher Dashboard in fullscreen mode
- AND display:
  - ASCII logo banner and application title
  - List of recent novels located in the configured root directory (with novel title, chapter count, and last modified date)
  - Action menu listing: `[c]` Continue Last Project, `[n]` New Novel, `[o]` Open Other Folder, `[l]` Configure LLM/AI, `[d]` Set Root Directory, `[q]` Quit.

#### Scenario: Navigating recent novels list
- GIVEN the Launcher Dashboard is displayed with multiple recent novels
- WHEN the user presses `Up`/`Down` arrow keys (or `k`/`j` when no input is focused)
- THEN the selected novel highlight MUST move up or down accordingly
- AND pressing `Enter` on a selected novel MUST load that novel and transition to the Editor view.

#### Scenario: Action menu keybindings
- GIVEN the Launcher Dashboard is active and no modal dialog is open
- WHEN the user presses `c`
- THEN the system MUST trigger the "Continue Last Project" action.
- WHEN the user presses `n`
- THEN the system MUST open the "New Novel" modal dialog.
- WHEN the user presses `l`
- THEN the system MUST open the "LLM Configuration" view.
- WHEN the user presses `d`
- THEN the system MUST open the "Set Root Directory" modal dialog.
- WHEN the user presses `o`
- THEN the system MUST open the folder selector modal or prompt.

#### Scenario: Quitting from Launcher
- GIVEN the Launcher Dashboard is active and no modal dialog is open
- WHEN the user presses `q` or `Ctrl+C`
- THEN the application MUST terminate cleanly and exit.

---

### Requirement: Global Configuration and Multi-Project Directory Structure

The system MUST manage a centralized JSON configuration file at `~/.config/novel-tui/config.json` and enforce a standardized directory layout for all novel projects under a configurable root path (`~/Novelas/` by default).

#### Scenario: Loading or initializing default global configuration
- GIVEN no existing configuration file at `~/.config/novel-tui/config.json`
- WHEN the application starts
- THEN the system MUST create the configuration directory `~/.config/novel-tui/` if missing
- AND write a default `config.json` containing:
  - `root_dir`: `~/Novelas` (expanded to absolute user home path)
  - `recent_novels`: empty list `[]`
  - `llm`: default LLM parameters (Provider: "ollama", BaseURL: "http://localhost:11434", Model: "qwen2.5:7b", Temperature: 0.7, GenrePrompts: default map)
- AND create the default novel root directory `~/Novelas/` if it does not already exist.

#### Scenario: Standardized project workspace structure
- GIVEN a novel named `<novel_name>` created or managed by `novel-tui`
- WHEN the novel workspace is scaffolded or loaded on disk
- THEN the project directory layout MUST strictly adhere to:
  - `<root_dir>/<novel_name>/`
  - `<root_dir>/<novel_name>/capitulos/` (directory holding `.txt` chapter files)
  - `<root_dir>/<novel_name>/personajes.json` (JSON file holding character and lore entries)
  - `<root_dir>/<novel_name>/notas.txt` (plain text scratchpad for author notes)

#### Scenario: Discovering existing novels in root directory
- GIVEN the configured root directory contains one or more subdirectories containing a `capitulos/` folder or valid project files
- WHEN the Launcher initializes or refreshes
- THEN the system MUST scan the root directory and populate the recent novels list with project names, counting `.txt` files inside each project's `capitulos/` directory.

---

### Requirement: New Novel Modal Dialog and Workspace Scaffolding

The system MUST provide a modal dialog to input a novel title, validate the input, automatically scaffold the project folder structure with an initial chapter, and transition directly into the Editor view.

#### Scenario: Creating a new novel with valid title
- GIVEN the New Novel modal dialog is open with an active text input
- WHEN the user types a valid title "Mi Primera Novela" and presses `Enter`
- THEN the system MUST sanitize the title into a safe folder name (e.g. `Mi_Primera_Novela` or `Mi Primera Novela`)
- AND create directory `<root_dir>/<folder_name>/capitulos/`
- AND create starter files:
  - `<root_dir>/<folder_name>/capitulos/01_capitulo_1.txt` with default starter header/content
  - `<root_dir>/<folder_name>/personajes.json` with an empty JSON array `[]`
  - `<root_dir>/<folder_name>/notas.txt` with empty or template note text
- AND update `config.json` to place this novel at the top of `recent_novels`
- AND close the modal dialog and transition immediately to the Editor view with `01_capitulo_1.txt` loaded.

#### Scenario: Validating empty or whitespace title
- GIVEN the New Novel modal dialog is open
- WHEN the user leaves the input empty or enters only whitespace and presses `Enter`
- THEN the system MUST reject the submission
- AND display an inline error message indicating that the novel title cannot be empty
- AND keep the modal dialog open with focus on the text input.

#### Scenario: Handling directory name collision
- GIVEN a project folder with the same target folder name already exists in `<root_dir>`
- WHEN the user submits the novel creation form
- THEN the system MUST NOT overwrite or delete the existing directory
- AND display a clear error message that a novel with this name already exists
- AND allow the user to modify the title or cancel.

#### Scenario: Cancelling New Novel modal
- GIVEN the New Novel modal dialog is open
- WHEN the user presses `Esc`
- THEN the system MUST close the modal dialog
- AND restore full focus to the Launcher Dashboard without creating any files on disk.

---

### Requirement: Continue Last Accessed Project

The system MUST allow instant access to the most recently opened novel project via the `[c]` action shortcut.

#### Scenario: Continuing when recent project exists
- GIVEN `config.json` has at least one entry in `recent_novels` with an existing path on disk
- WHEN the user presses `c` in the Launcher Dashboard
- THEN the system MUST load the project located at the top of the `recent_novels` list
- AND load its most recent or first chapter into the Editor
- AND transition the UI state from Launcher to Editor view.

#### Scenario: Continuing when no recent projects exist
- GIVEN `config.json` has an empty `recent_novels` list or referenced directories no longer exist
- WHEN the user presses `c` in the Launcher Dashboard
- THEN the system MUST display a non-blocking notification or status message indicating that no recent project was found
- AND remain on the Launcher Dashboard.

---

### Requirement: LLM Settings Configuration View

The system MUST provide a settings screen to inspect and modify local LLM parameters and genre prompt templates, persisting changes to `~/.config/novel-tui/config.json`.

#### Scenario: Viewing LLM configuration fields
- GIVEN the user navigates to the LLM Configuration view (`[l]`)
- WHEN the view renders
- THEN the system MUST display editable fields for:
  - `Provider` (default: "ollama")
  - `Base URL` (default: "http://localhost:11434")
  - `Model Name` (default: "qwen2.5:7b")
  - `Temperature` (numeric range between 0.0 and 2.0, default: 0.7)
  - `Genre Prompt Templates` (selectable/editable prompts for genres such as Fantasy, Sci-Fi, Mystery, Romance).

#### Scenario: Saving modified LLM settings
- GIVEN the user modifies one or more LLM configuration fields (e.g. changes Model Name to "mistral:latest" and Temperature to 0.8)
- WHEN the user confirms save via `Ctrl+S` or `Enter` on the Save action
- THEN the system MUST validate the field values (URL format, valid temperature range)
- AND update the `llm` block in `~/.config/novel-tui/config.json`
- AND display a confirmation indicator
- AND allow returning to the Launcher Dashboard with updated settings active.

#### Scenario: Cancelling LLM configuration edits
- GIVEN the user has made uncommitted changes in the LLM Configuration view
- WHEN the user presses `Esc`
- THEN the system MUST discard unsaved changes and return to the Launcher Dashboard.

---

### Requirement: Root Directory Settings Modal

The system MUST provide a modal dialog allowing the user to inspect and change the base directory where novels are stored and discovered.

#### Scenario: Changing the novel root directory
- GIVEN the user triggers the "Set Root Directory" action (`[d]`) from the Launcher Dashboard
- WHEN the modal dialog appears
- THEN it MUST display a text input pre-filled with the current `root_dir` path
- WHEN the user enters a new valid directory path (e.g. `~/Documentos/MisNovelas`) and presses `Enter`
- THEN the system MUST expand tilde (`~`) to the user home directory
- AND create the directory if it does not already exist
- AND update `root_dir` in `~/.config/novel-tui/config.json`
- AND refresh the Launcher's recent novels list by scanning the new root directory
- AND close the modal dialog.

#### Scenario: Cancelling root directory change
- GIVEN the Set Root Directory modal dialog is open
- WHEN the user presses `Esc`
- THEN the modal dialog MUST close without altering the configured `root_dir`.

---

### Requirement: Chapter Creation Modal in Editor

The system MUST provide a modal dialog inside the Editor view triggered by `Ctrl+N` or `n` (when chapter list has focus) to name and create new chapters without accidental duplicates.

#### Scenario: Opening chapter creation modal in Editor
- GIVEN the user is in the Editor view
- WHEN the user presses `Ctrl+N` (or `n` while the sidebar chapter list is focused)
- THEN the system MUST open a centered TextInput modal dialog titled "Nuevo Capítulo"
- AND focus MUST move to the modal text input.

#### Scenario: Creating chapter with formatted number and slug
- GIVEN the novel currently contains existing chapters `01_prologo.txt` and `02_el_despertar.txt`
- WHEN the user enters title "Encuentro en la Taberna" in the modal and presses `Enter`
- THEN the system MUST calculate the next sequence number (`03`)
- AND generate the filename `03_encuentro_en_la_taberna.txt` inside `<novel_root>/capitulos/`
- AND write an initial chapter header into the file
- AND add the new chapter to the sidebar chapter list
- AND select and load the new chapter in the Center Editor
- AND transfer focus to the Center Editor.

#### Scenario: Cancelling chapter creation in Editor
- GIVEN the Chapter Creation modal is visible in the Editor view
- WHEN the user presses `Esc`
- THEN the modal MUST close immediately
- AND no new file MUST be created on disk
- AND focus MUST return to the previous active panel (Sidebar or Editor).

---

### Requirement: Spellchecker Domain and Interface Contracts

The system MUST define domain interfaces for spellchecking and dictionary management decoupled from concrete CGO or external process implementations.

#### Scenario: Spellchecker interface definition
- GIVEN the domain package `internal/domain`
- WHEN the spellchecker interface is declared
- THEN it MUST define:
  ```go
  type Spellchecker interface {
      Check(word string) bool
      Suggestions(word string) []string
  }
  ```
- AND `Check` MUST return `true` if the given word is valid according to the active dictionary, and `false` otherwise.
- AND `Suggestions` MUST return an ordered slice of recommended candidate words for a misspelled word.

#### Scenario: DictionaryManager interface definition
- GIVEN the domain package `internal/domain`
- WHEN the dictionary manager interface is declared
- THEN it MUST define:
  ```go
  type DictionaryManager interface {
      LoadDictionary(affPath, dicPath string) error
      AddCustomWord(word string) error
      AvailableDictionaries() []string
  }
  ```
- AND `LoadDictionary` MUST load `.aff` (affix) and `.dic` (wordlist) files into memory or the underlying engine.
- AND `AddCustomWord` MUST persist user-approved words to the author's local custom dictionary.

---

### Requirement: Fullscreen AltScreen Terminal Lifecycle

The system MUST initialize Bubble Tea with `tea.WithAltScreen()` ensuring that all views (Launcher, Editor, Settings) render cleanly without visual artifacts or leaking buffer content into the terminal scrollback history upon exit.

#### Scenario: Clean application startup and shutdown
- GIVEN the application is executed from a terminal emulator
- WHEN `tea.NewProgram` is invoked
- THEN the program options MUST include `tea.WithAltScreen()`
- AND when the application receives `tea.Quit` or handles termination
- THEN the terminal MUST exit alt-screen mode, restore the previous terminal buffer, and leave the terminal cursor in a clean state without residue.

#### Scenario: Seamless view switching between Launcher and Editor
- GIVEN the user transitions between Launcher Dashboard, Config views, and Editor view
- WHEN view state transitions occur
- THEN the entire terminal screen area MUST be re-rendered cleanly based on `tea.WindowSizeMsg` dimensions
- AND no ghost characters or overlapping frames from previous views MUST remain visible.

---

### Requirement: Genre-Adaptive AI Writing Assistant and Agent Mode

The system MUST support an integrated AI Writing Assistant with two operational modes (Assistant Chat and Autonomous Agent Co-Writer) that dynamically adapts its persona, vocabulary, pacing advice, and critique style based on the author's selected genre(s) for the novel.

#### Scenario: Interactive AI Writing Chat Panel
- GIVEN the Editor view is open with an active chapter
- WHEN the user toggles the AI Chat drawer or panel (e.g. `Ctrl+A` or mouse click)
- THEN a dedicated AI interaction panel MUST open alongside the editor without disrupting active writing
- AND the user MUST be able to ask questions about plot structure, character consistency, scene pacing, and dialogue flow.

#### Scenario: Genre-Adaptive Persona and Prompting
- GIVEN the novel project has configured genre tags (e.g., Fantasía, Sci-Fi, Romance, Suspenso/Misterio, Peleas/Shonen, Psicológico)
- WHEN the AI generates suggestions, feedback, or text continuations
- THEN the system MUST inject specialized system instructions tailored to those genres (e.g., worldbuilding coherence and magic rules for Fantasy, high tension and sensory beats for Thriller, emotional depth and dialogue subtext for Romance, dynamic choreography and momentum for Action)
- AND the AI MUST adopt the tone of an expert editor specialized in that specific literary domain.

#### Scenario: Agent Mode (Context-Aware Autonomous Co-Writer)
- GIVEN the AI panel is set to "Agent Mode"
- WHEN the user requests an agent task (e.g., "Analizar coherencia del personaje Kuno", "Generar borrador de escena", "Sugerir 3 giros argumentales para el capítulo actual")
- THEN the Agent MUST autonomously read the relevant project context:
  - Current and previous chapter contents in `capitulos/*.txt`
  - Character bios and lore in `personajes.json`
  - Author outlines and scratchpad notes in `notas.txt`
- AND produce structured, contextually aware responses incorporating historical lore and character traits.

---

### Requirement: Full Mouse Support and Interactive UI Navigation

The system MUST support terminal mouse events (`tea.WithMouseCellMotion()` or standard click/scroll tracking) across all views (Launcher, Editor, Settings, and Modals) alongside existing keyboard shortcuts.

#### Scenario: Mouse clicking in Sidebar and Launcher
- GIVEN the Launcher or Editor sidebar is visible
- WHEN the user clicks with the left mouse button on a novel card or chapter list item
- THEN the clicked item MUST be selected immediately
- AND a double-click (or single click on action buttons) MUST trigger opening the novel or loading the chapter into the editor.

#### Scenario: Mouse clicking on Sidebar Tabs
- GIVEN the Sidebar is displaying the tab header (`[Chapters]` / `[Lore / Characters]`)
- WHEN the user clicks on a tab title
- THEN the active tab MUST switch to the clicked tab without requiring keyboard hotkeys.

#### Scenario: Mouse scroll wheel navigation
- GIVEN the Editor textarea, Chapter list, or Recent Novels list has scrollable overflow
- WHEN the user rotates the mouse wheel up or down over the respective component
- THEN the view MUST scroll content smoothly up or down.

#### Scenario: Mouse positioning in Text Editor
- GIVEN the central Text Editor is focused
- WHEN the user clicks on a specific character position inside the text area
- THEN the text cursor MUST move to the clicked line and column.

