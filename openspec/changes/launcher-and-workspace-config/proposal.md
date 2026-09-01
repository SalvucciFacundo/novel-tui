# SDD Proposal: Launcher & Workspace Config

## Intent
The goal of this change is to transform `novel-tui` from a single-directory terminal editor into a comprehensive, multi-project novel writing environment. This introduces a home dashboard (launcher) that serves as the entry point for recent projects, global settings, LLM configuration, and workspace organization. Additionally, it establishes the foundation for future AI integrations and spellchecking support, while refining the terminal UX.

## Scope
### In-Scope
1. **Home Dashboard / Launcher TUI:**
   - Display on startup when no specific directory is provided.
   - ASCII Logo banner and title.
   - List of recent novels with chapter counts and navigation.
   - Main actions: `[c]` Continue Last Project, `[n]` New Novel, `[o]` Open Other Folder, `[l]` Configure LLM/AI, `[d]` Set Root Directory, `[q]` Quit.
2. **Configurable Root Directory & Organization:**
   - Central configuration file in `~/.config/novel-tui/config.json`.
   - Default root path: `~/Novelas`.
   - Standardized multi-novel structure:
     - `capitulos/` (containing `.txt` chapters)
     - `personajes.json`
     - `notas.txt`
3. **LLM Configuration Screen:**
   - Manage settings for local AI (Provider: Ollama, Endpoint URL, Model name, Temperature).
   - Configure genre prompt templates.
4. **Terminal Display Fix:**
   - Ensure `tea.WithAltScreen()` is properly activated in Bubble Tea initialization and that all views (Launcher, Editor, Config) transition smoothly without leaking terminal history.
5. **Chapter Creation Modal:**
   - Triggered by `Ctrl+N` or `n`.
   - Input prompt modal for the Chapter Title.
   - Prevention of duplicate or rapid creation.
6. **Spellchecker Foundation:**
   - Domain models and interfaces defining the service architecture for Hunspell/dictionary (`.aff`/`.dic`) spellchecking, decoupled from the UI.

### Out-of-Scope
- Implementing the actual LLM chat/generation logic (this phase only builds the configuration UI and storage).
- Implementing the actual Hunspell CGO/binary bindings (only domain/service interfaces are designed).
- Cloud sync or external database integrations for projects.

## Affected Areas
- **Initialization & CLI (`cmd/novel-tui/main.go`):** Modification to launch the dashboard instead of directly opening the current directory, loading global config.
- **UI Models (`internal/ui/model/root.go`):** Introduction of states/views for Dashboard, Editor, and Config. State machine transitions will need refactoring.
- **Services (`internal/service/workspace.go`):** Needs expansion to handle multi-project discovery, configuration loading/saving (`~/.config/novel-tui/config.json`), and project scaffolding.
- **Domain (`internal/domain/`):** New structs for AppConfig, LLMConfig, and Spellchecker interfaces.
- **Repositories (`internal/repository/`):** Adaptation to standard `.txt` saving instead of previous defaults, if necessary.

## Risks & Edge Cases
- **Risk:** Existing workspaces might not map to the new `~/Novelas/<Project>` structure, causing backward compatibility issues.
- **Risk:** Terminal state transitions between the Launcher and Editor could cause UI artifacts if `tea` commands aren't properly sequenced.
- **Edge Case:** Missing permissions to create `~/.config/novel-tui/` or `~/Novelas/`.
- **Edge Case:** Duplicate novel names or invalid folder characters during "New Novel" creation.

## Rollback Plan
Since the project is in early development, rollback involves reverting the UI state machine to immediately launch the editor, and ignoring the `~/.config` file. Commits should be isolated so the entire `launcher-and-workspace-config` branch can be dropped or reverted without impacting the core editor text entry logic.

## Success Criteria
1. Running `novel-tui` without arguments displays the new ASCII Home Dashboard.
2. The user can create a new novel, which automatically scaffolds `~/Novelas/<Name>/` with the correct subdirectories and files.
3. Chapters are created through a modal prompt and saved explicitly as `.txt`.
4. Global configuration is persisted to `~/.config/novel-tui/config.json`.
5. The LLM configuration screen allows updating and saving the endpoint, model, and templates.
6. The terminal clears cleanly upon exit without leaving editor content in the history.

## Proposal Question Round
*To ensure the product requirements match the architectural approach, please review the following assumptions. If everything looks correct, we can proceed to Spec.*

1. **Migration / Existing Data:** If a user runs `novel-tui` in a directory that already has files but doesn't follow the new `~/Novelas/` structure, should we prompt to migrate them, or just treat it as an ad-hoc workspace?
2. **Recent Projects:** How many recent projects should be stored in the config, and should we sort them strictly by last accessed time?
3. **Chapter Naming Convention:** The spec mentions `01_capitulo.txt`. Should the app automatically prefix chapter files with sequential numbers, or does the user input the exact file name in the modal?
4. **Spellchecker Foundation:** Is the expectation to eventually use CGO (like `godoc.org/github.com/microcosm-cc/bluemonday` or similar hunspell wrappers), or an external command-line process for hunspell? (This helps define the async nature of the interface).