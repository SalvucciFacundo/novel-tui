# Implementation Tasks: UI Redesign, Navbar, and Polish

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 350 - 500 lines |
| 400-line budget risk | Medium |
| Chained PRs recommended | No |
| Suggested split | single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Medium

---

## Tasks

### Phase 1: Top Navigation Bar (Navbar) Implementation
- [x] Create `internal/ui/components/navbar.go` implementing `Navbar` struct, breadcrumbs rendering (`📖 Novel › 📑 Chapter`), and interactive action pills (`[← Inicio (Ctrl+H)]`, `[1: Capítulos]`, `[2: Personajes]`, `[3: Notas]`, `[🤖 Asistente IA (Ctrl+A)]`). <!-- sdd-owner: implementation -->
- [x] Add mouse click hit-testing in `Navbar.Update` to emit custom navigation / action messages (`ReturnHomeMsg`, `ChangeSidebarTabMsg`, `ToggleAIChatMsg`). <!-- sdd-owner: implementation -->

### Phase 2: Footer Status Bar Enhancement
- [x] Update `internal/ui/components/statusbar.go` to render left-aligned command badges (`[Ctrl+H: Inicio]`, `[Ctrl+A: IA]`, `[Ctrl+S: Guardar]`, `[Ctrl+N: Nuevo Cap]`, `[Tab: Panel]`). <!-- sdd-owner: implementation -->
- [x] Add right-aligned dynamic metrics (`{Words} palabras`, `{Chars} caracteres`, `~{Time} min`) and save status indicator (`[Guardado]` vs `[Modificado*]`). <!-- sdd-owner: implementation -->

### Phase 3: Sidebar 3-Tab Management & Notes File Integration
- [x] Refactor `internal/ui/components/sidebar.go` to support 3 tabs (`TabChapters`, `TabCharacters`, `TabNotes`) with hotkeys `1`, `2`, `3` and tab header rendering. <!-- sdd-owner: implementation -->
- [x] Wire up loading and saving of `notas.txt` from the novel's root directory in `TabNotes`. <!-- sdd-owner: implementation -->

### Phase 4: Modal Polish, Editor Reset, and Scoped Keybindings
- [x] Fix modal background styling in `internal/ui/components/modal.go` by applying uniform Lip Gloss card background styling to inner elements and containers to eliminate black strips. <!-- sdd-owner: implementation -->
- [x] Update `internal/ui/components/editor.go` to reset cursor position (`(0, 0)`) and viewport scroll to top on chapter load. <!-- sdd-owner: implementation -->
- [x] Update `internal/ui/model/root.go` to assemble top `Navbar`, recalculate layout height (`height - navbarHeight - statusbarHeight`), handle `Ctrl+H` globally, and scope `Ctrl+N` so it does not trigger when focus is in the AI Chat Drawer. <!-- sdd-owner: implementation -->

### Phase 5: Testing and Verification
- [x] Update unit and component test suites in `internal/ui/components/components_test.go` and `internal/ui/model/root_test.go` to cover Navbar, 3-tab Sidebar, and scoped keybindings. <!-- sdd-owner: implementation -->
- [x] Run test suite and verify UI layout rendering and functionality. <!-- sdd-owner: implementation -->
