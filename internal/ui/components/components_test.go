package components_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/repository"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/components"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

func TestNavbarComponent(t *testing.T) {
	styles := theme.DefaultStyles
	nav := components.NewNavbarModel(styles)
	nav.SetWidth(140)
	nav.SetNovelTitle("Cien Años de Soledad")
	nav.SetChapterTitle("Capítulo 1")

	view := nav.View()
	if !strings.Contains(view, "Cien Años de Soledad") {
		t.Errorf("expected novel title in navbar view: %s", view)
	}
	if !strings.Contains(view, "Capítulo 1") {
		t.Errorf("expected chapter title in navbar view: %s", view)
	}
	if !strings.Contains(view, "Inicio") {
		t.Errorf("expected Inicio pill in navbar: %s", view)
	}
	if !strings.Contains(view, "Asistente IA") {
		t.Errorf("expected Asistente IA pill in navbar: %s", view)
	}

	// Fallback breadcrumbs when chapter is empty
	navEmpty := components.NewNavbarModel(styles)
	navEmpty.SetWidth(140)
	navEmpty.SetNovelTitle("Mi Novela")
	viewEmpty := navEmpty.View()
	if !strings.Contains(viewEmpty, "Ningún capítulo seleccionado") {
		t.Errorf("expected fallback chapter text in navbar: %s", viewEmpty)
	}

	// Test mouse click hit-testing
	// 1. Click Inicio pill (at X = 5)
	_, cmd := nav.Update(tea.MouseMsg{
		X:    5,
		Y:    0,
		Type: tea.MouseLeft,
	})
	if cmd == nil {
		t.Fatalf("expected command on clicking Inicio pill")
	}
	msg := cmd()
	viewMsg, ok := msg.(messages.ChangeViewMsg)
	if !ok || viewMsg.View != messages.ViewStateLauncher {
		t.Errorf("expected ChangeViewMsg(Launcher), got: %+v", msg)
	}

	// 2. Click AI Assistant pill (at X = 125)
	_, cmd = nav.Update(tea.MouseMsg{
		X:    125,
		Y:    0,
		Type: tea.MouseLeft,
	})
	if cmd == nil {
		t.Fatalf("expected command on clicking AI pill")
	}
	msg = cmd()
	if _, ok := msg.(messages.ToggleChatDrawerMsg); !ok {
		t.Errorf("expected ToggleChatDrawerMsg, got: %+v", msg)
	}

	// 3. Click Tab 1: Capítulos (at X = 70)
	_, cmd = nav.Update(tea.MouseMsg{
		X:    70,
		Y:    0,
		Type: tea.MouseLeft,
	})
	if cmd == nil {
		t.Fatalf("expected command on clicking Tab 1 pill")
	}
	msg = cmd()
	tabMsg, ok := msg.(messages.SelectSidebarTabMsg)
	if !ok || tabMsg.Tab != 0 {
		t.Errorf("expected SelectSidebarTabMsg(0), got: %+v", msg)
	}

	// 4. Click Tab 2: Personajes (at X = 90)
	_, cmd = nav.Update(tea.MouseMsg{
		X:    90,
		Y:    0,
		Type: tea.MouseLeft,
	})
	if cmd == nil {
		t.Fatalf("expected command on clicking Tab 2 pill")
	}
	msg = cmd()
	tabMsg, ok = msg.(messages.SelectSidebarTabMsg)
	if !ok || tabMsg.Tab != 1 {
		t.Errorf("expected SelectSidebarTabMsg(1), got: %+v", msg)
	}

	// 5. Click Tab 3: Notas (at X = 105)
	_, cmd = nav.Update(tea.MouseMsg{
		X:    105,
		Y:    0,
		Type: tea.MouseLeft,
	})
	if cmd == nil {
		t.Fatalf("expected command on clicking Tab 3 pill")
	}
	msg = cmd()
	tabMsg, ok = msg.(messages.SelectSidebarTabMsg)
	if !ok || tabMsg.Tab != 2 {
		t.Errorf("expected SelectSidebarTabMsg(2), got: %+v", msg)
	}
}

func TestStatusBarComponent(t *testing.T) {
	styles := theme.DefaultStyles
	sb := components.NewStatusBarModel(styles)
	sb.SetWidth(100)

	// Clean chapter loaded
	sb, _ = sb.Update(messages.ChapterSelectedMsg{
		Chapter: domain.Chapter{
			ID:      "chap-1",
			Title:   "Capítulo 1",
			Content: "Una palabra dos palabras tres palabras cuatro palabras",
		},
	})

	view := sb.View()
	if !strings.Contains(view, "Ctrl+H: Inicio") {
		t.Errorf("expected Ctrl+H command badge in statusbar: %s", view)
	}
	if !strings.Contains(view, "Ctrl+A: IA") {
		t.Errorf("expected Ctrl+A command badge in statusbar: %s", view)
	}
	if !strings.Contains(view, "Guardado") {
		t.Errorf("expected [Guardado] pill in statusbar: %s", view)
	}
	if !strings.Contains(view, "palabras") || !strings.Contains(view, "caracteres") {
		t.Errorf("expected metrics formatting in statusbar: %s", view)
	}

	// Dirty chapter text changed
	sb, _ = sb.Update(messages.TextChangedMsg{
		ChapterID: "chap-1",
		Content:   "Modificado",
		Metrics: domain.EditorMetrics{
			WordCount:   10,
			CharCount:   50,
			ReadingTime: 1,
			IsDirty:     true,
		},
	})
	dirtyView := sb.View()
	if !strings.Contains(dirtyView, "Modificado*") {
		t.Errorf("expected [Modificado*] pill on dirty state: %s", dirtyView)
	}
}

func TestSidebar3TabsAndNotes(t *testing.T) {
	tempDir := t.TempDir()
	chapRepo, _ := repository.NewFileChapterRepository(tempDir)
	charRepo := repository.NewFileCharacterRepository(tempDir)
	_ = charRepo.SaveAll([]domain.Character{{ID: "c1", Name: "Aurelio", Role: "Mago"}})

	// Create a notes file in the novel root
	notesPath := filepath.Join(tempDir, "notas.txt")
	_ = os.WriteFile(notesPath, []byte("Lista de ideas para la trama"), 0644)

	styles := theme.DefaultStyles
	sidebar := components.NewSidebarModel(chapRepo, charRepo, styles)
	sidebar.SetSize(40, 20)
	sidebar.SetNovelPath(tempDir)
	sidebar, _ = sidebar.Update(messages.FocusMsg{Target: messages.FocusSidebar})

	// 1. Initial tab is TabChapters
	if sidebar.ActiveTab != components.TabChapters {
		t.Errorf("expected initial tab TabChapters, got %v", sidebar.ActiveTab)
	}

	// 2. Switch to Tab 2: Personajes via key '2'
	sidebar, _ = sidebar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	if sidebar.ActiveTab != components.TabCharacters {
		t.Errorf("expected TabCharacters on pressing '2', got %v", sidebar.ActiveTab)
	}

	// 3. Switch to Tab 3: Notas via key '3'
	sidebar, _ = sidebar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
	if sidebar.ActiveTab != components.TabNotes {
		t.Errorf("expected TabNotes on pressing '3', got %v", sidebar.ActiveTab)
	}

	// Check notes loaded
	if !strings.Contains(sidebar.NotesValue(), "Lista de ideas para la trama") {
		t.Errorf("expected notes loaded from notas.txt, got %s", sidebar.NotesValue())
	}

	// 4. Edit notes and save
	sidebar, _ = sidebar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" extra")})
	sidebar, _ = sidebar.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	savedData, _ := os.ReadFile(notesPath)
	if !strings.Contains(string(savedData), "extra") {
		t.Errorf("expected saved notes on disk to contain 'extra', got: %s", string(savedData))
	}

	// 5. SelectSidebarTabMsg switching
	sidebar, _ = sidebar.Update(messages.SelectSidebarTabMsg{Tab: 0})
	if sidebar.ActiveTab != components.TabChapters {
		t.Errorf("expected TabChapters after SelectSidebarTabMsg(0), got %v", sidebar.ActiveTab)
	}
}

func TestEditorCursorReset(t *testing.T) {
	styles := theme.DefaultStyles
	editor := components.NewEditorModel(styles)
	editor.SetSize(60, 20)

	// Load long chapter
	longContent := strings.Repeat("Linea de texto\n", 50)
	editor, _ = editor.Update(messages.ChapterSelectedMsg{
		Chapter: domain.Chapter{
			ID:      "chap-1",
			Title:   "Capítulo 1",
			Content: longContent,
		},
	})

	if editor.Value() != longContent {
		t.Errorf("expected editor content to match loaded chapter")
	}
}

func TestModalComponent(t *testing.T) {
	styles := theme.DefaultStyles
	modal := components.NewModalModel(styles)
	modal.SetSize(80, 24)

	// Initially inactive
	if modal.Active {
		t.Errorf("modal should not be active initially")
	}
	if modal.View() != "" {
		t.Errorf("modal view should be empty when inactive")
	}

	// Show modal
	modal.Show(messages.ModalPurposeNewNovel, "Nueva Novela", "Título:", "")
	if !modal.Active {
		t.Errorf("modal should be active after Show()")
	}

	view := modal.View()
	if !strings.Contains(view, "Nueva Novela") {
		t.Errorf("expected title in modal view, got: %s", view)
	}

	// Submit empty input -> should trigger error
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if modal.ErrorMsg == "" {
		t.Errorf("expected validation error on empty submit")
	}

	// Type input and submit
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Mi Novela")})
	modal, cmd := modal.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if modal.Active {
		t.Errorf("modal should be closed after valid submission")
	}
	if cmd == nil {
		t.Fatalf("expected command emitted on valid modal submit")
	}
	msg := cmd()
	createMsg, ok := msg.(messages.CreateNovelMsg)
	if !ok || createMsg.Title != "Mi Novela" {
		t.Errorf("unexpected message from modal submit: %+v", msg)
	}
}

func TestLauncherComponent(t *testing.T) {
	styles := theme.DefaultStyles
	launcher := components.NewLauncherModel(styles)
	launcher.SetSize(100, 30)

	novels := []domain.NovelMetadata{
		{
			Title:        "Novela de Prueba",
			AbsolutePath: "/tmp/novela-prueba",
			ChapterCount: 3,
			LastModified: time.Now(),
		},
	}
	launcher.SetNovels(novels)

	view := launcher.View()
	if !strings.Contains(view, "Novela de Prueba") {
		t.Errorf("expected novel title in launcher view: %s", view)
	}

	// Press 'n' -> should emit ShowModalMsg
	_, cmd := launcher.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if cmd == nil {
		t.Fatalf("expected command on pressing 'n'")
	}
	msg := cmd()
	showMsg, ok := msg.(messages.ShowModalMsg)
	if !ok || showMsg.Purpose != messages.ModalPurposeNewNovel {
		t.Errorf("unexpected message on 'n': %+v", msg)
	}

	// Press 'l' -> should emit ChangeViewMsg for LLM config
	_, cmd = launcher.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if cmd == nil {
		t.Fatalf("expected command on pressing 'l'")
	}
	msg = cmd()
	viewMsg, ok := msg.(messages.ChangeViewMsg)
	if !ok || viewMsg.View != messages.ViewStateLLMConfig {
		t.Errorf("unexpected view change message: %+v", msg)
	}
}

func TestLLMConfigComponent(t *testing.T) {
	styles := theme.DefaultStyles
	llmCfg := components.NewLLMConfigModel(styles)
	llmCfg.SetSize(100, 30)

	view := llmCfg.View()
	if !strings.Contains(view, "Configuración de Inteligencia Artificial") {
		t.Errorf("expected LLM config header in view: %s", view)
	}

	// Press 'Esc' -> should emit ChangeViewMsg for Launcher
	_, cmd := llmCfg.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatalf("expected command on pressing 'Esc'")
	}
	msg := cmd()
	viewMsg, ok := msg.(messages.ChangeViewMsg)
	if !ok || viewMsg.View != messages.ViewStateLauncher {
		t.Errorf("unexpected view change message on Esc: %+v", msg)
	}

	// Press 'Ctrl+S' -> should emit SaveLLMConfigMsg
	_, cmd = llmCfg.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd == nil {
		t.Fatalf("expected command on pressing 'Ctrl+S'")
	}
	msg = cmd()
	saveMsg, ok := msg.(messages.SaveLLMConfigMsg)
	if !ok || saveMsg.Config.Provider != "ollama" {
		t.Errorf("unexpected save message: %+v", msg)
	}
}

func TestChatDrawerComponent(t *testing.T) {
	tempDir := t.TempDir()
	sessionRepo := repository.NewFileChatSessionRepository()
	styles := theme.DefaultStyles

	drawer := components.NewChatDrawerModel(sessionRepo, styles)
	drawer.SetNovelDir(tempDir)
	drawer.SetDimensions(40, 24)

	// 1. Initial State
	if drawer.EffortLevel() != domain.EffortMedium {
		t.Errorf("expected default effort Medium, got %s", drawer.EffortLevel())
	}

	// 2. Effort Level Cycling
	drawer.CycleEffortLevel()
	if drawer.EffortLevel() != domain.EffortHigh {
		t.Errorf("expected effort High after cycle, got %s", drawer.EffortLevel())
	}
	drawer.CycleEffortLevel()
	if drawer.EffortLevel() != domain.EffortLow {
		t.Errorf("expected effort Low after cycle, got %s", drawer.EffortLevel())
	}

	// 3. Focus Drawer
	drawer, _ = drawer.Update(messages.FocusMsg{Target: messages.FocusChat})
	if !drawer.Focused {
		t.Errorf("expected drawer to be focused")
	}

	// 4. Token Streaming
	drawer, _ = drawer.Update(messages.TokenReceivedMsg{Content: "Hola"})
	drawer, _ = drawer.Update(messages.TokenReceivedMsg{Content: " mundo"})
	drawer, _ = drawer.Update(messages.StreamFinishedMsg{})

	sess := drawer.ActiveSession()
	if len(sess.Messages) == 0 {
		t.Fatalf("expected messages in active session")
	}
	lastMsg := sess.Messages[len(sess.Messages)-1]
	if lastMsg.Role != "assistant" || lastMsg.Content != "Hola mundo" {
		t.Errorf("expected assistant message 'Hola mundo', got '%s'", lastMsg.Content)
	}

	// 5. View rendering
	view := drawer.View()
	if !strings.Contains(view, "Asistente IA") && !strings.Contains(view, "Nueva Conversación") {
		t.Errorf("expected drawer view to contain header, got:\n%s", view)
	}
	if !strings.Contains(view, "Hola mundo") {
		t.Errorf("expected drawer view to contain 'Hola mundo'")
	}
}

func TestNavbarBrainTab(t *testing.T) {
	styles := theme.DefaultStyles
	nav := components.NewNavbarModel(styles)
	nav.SetWidth(140)
	nav.SetNovelTitle("Mi Novela")

	view := nav.View()
	if !strings.Contains(view, "4: Brain") {
		t.Errorf("expected '4: Brain' pill in navbar view, got: %s", view)
	}

	// Click 4: Brain pill (around X = 100 in 140-width navbar)
	// Let's test clicking through mouse message
	zones := nav.View()
	_ = zones

	// Simulate clicking tab 4 pill
	clicked := false
	for x := 0; x < 140; x++ {
		_, cmd := nav.Update(tea.MouseMsg{
			X:    x,
			Y:    0,
			Type: tea.MouseLeft,
		})
		if cmd != nil {
			msg := cmd()
			if tabMsg, ok := msg.(messages.SelectSidebarTabMsg); ok && tabMsg.Tab == int(components.TabBrain) {
				clicked = true
				break
			}
		}
	}
	if !clicked {
		t.Errorf("expected clicking [4: Brain] pill in navbar to emit SelectSidebarTabMsg with TabBrain")
	}
}

func TestSidebarBrainTabAndFactManagement(t *testing.T) {
	tempDir := t.TempDir()
	chapRepo, _ := repository.NewFileChapterRepository(tempDir)
	charRepo := repository.NewFileCharacterRepository(tempDir)
	brainDbPath := filepath.Join(tempDir, "brain.db")
	brainRepo, err := repository.NewSQLiteBrainRepository(brainDbPath)
	if err != nil {
		t.Fatalf("failed to create brain repo: %v", err)
	}
	defer brainRepo.Close()

	ctx := context.Background()
	_ = brainRepo.SaveFact(ctx, domain.BrainFact{
		ID:      "fact-1",
		Topic:   "Personajes",
		Concept: "Kuno",
		Fact:    "Kuno perdió el brazo izquierdo en el capítulo 3",
		Type:    domain.FactTypeCharacter,
		Tags:    []string{"protagonista", "combate"},
	})
	_ = brainRepo.SaveFact(ctx, domain.BrainFact{
		ID:      "fact-2",
		Topic:   "Magia",
		Concept: "Espada del Sol",
		Fact:    "Forjada con fuego sagrado",
		Type:    domain.FactTypeLore,
		Tags:    []string{"artefacto"},
	})

	styles := theme.DefaultStyles
	sidebar := components.NewSidebarModel(chapRepo, charRepo, styles)
	sidebar.SetSize(50, 24)
	sidebar, _ = sidebar.Update(messages.FocusMsg{Target: messages.FocusSidebar})

	// Set brain repo and execute the returned cmd
	cmd := sidebar.SetBrainRepository(brainRepo)
	if cmd != nil {
		msg := cmd()
		sidebar, _ = sidebar.Update(msg)
	}

	if len(sidebar.BrainFacts) != 2 {
		t.Fatalf("expected 2 brain facts loaded in sidebar, got %d", len(sidebar.BrainFacts))
	}

	// 1. Switch directly to Tab 4: Brain via key '4'
	sidebar, _ = sidebar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})
	if sidebar.ActiveTab != components.TabBrain {
		t.Errorf("expected ActiveTab TabBrain on pressing '4', got %v", sidebar.ActiveTab)
	}

	// 2. View rendered content
	view := sidebar.View()
	if !strings.Contains(view, "Memoria Brain") {
		t.Errorf("expected 'Memoria Brain' header in TabBrain view, got: %s", view)
	}
	if !strings.Contains(view, "Kuno") || !strings.Contains(view, "Espada del Sol") {
		t.Errorf("expected facts concepts in TabBrain view, got: %s", view)
	}
	if !strings.Contains(view, "[d] Borrar hecho") {
		t.Errorf("expected delete hint '[d] Borrar hecho' in TabBrain view, got: %s", view)
	}

	// 3. Tab cycling with '[' and ']' across all 4 tabs
	// Current is TabBrain (3). Next tab ']' should wrap to TabChapters (0)
	sidebar, _ = sidebar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]")})
	if sidebar.ActiveTab != components.TabChapters {
		t.Errorf("expected TabChapters after cycling next from TabBrain, got %v", sidebar.ActiveTab)
	}
	// Prev tab '[' from TabChapters (0) should wrap to TabBrain (3)
	sidebar, _ = sidebar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("[")})
	if sidebar.ActiveTab != components.TabBrain {
		t.Errorf("expected TabBrain after cycling prev from TabChapters, got %v", sidebar.ActiveTab)
	}

	// 4. Navigation inside TabBrain: Down/j, Up/k
	sidebar.SelectedBrainFact = 0
	sidebar, _ = sidebar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if sidebar.SelectedBrainFact != 1 {
		t.Errorf("expected SelectedBrainFact == 1 after pressing 'j', got %d", sidebar.SelectedBrainFact)
	}
	sidebar, _ = sidebar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if sidebar.SelectedBrainFact != 0 {
		t.Errorf("expected SelectedBrainFact == 0 after pressing 'k', got %d", sidebar.SelectedBrainFact)
	}

	// 5. Delete selected fact via 'd'
	sidebar, delCmd := sidebar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	if delCmd != nil {
		msg := delCmd()
		sidebar, _ = sidebar.Update(msg)
	}

	// Check that facts decreased in repo and in sidebar
	facts, _ := brainRepo.ListRecentFacts(ctx, 100)
	if len(facts) != 1 {
		t.Errorf("expected 1 fact in brain repo after deletion, got %d", len(facts))
	}
	if len(sidebar.BrainFacts) != 1 {
		t.Errorf("expected 1 fact in sidebar after deletion, got %d", len(sidebar.BrainFacts))
	}

	// 6. Delete remaining fact to test empty state
	sidebar, delCmd2 := sidebar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	if delCmd2 != nil {
		msg := delCmd2()
		sidebar, _ = sidebar.Update(msg)
	}
	emptyView := sidebar.View()
	if !strings.Contains(emptyView, "Brain está activo y aprendiendo") {
		t.Errorf("expected empty state message in TabBrain view, got: %s", emptyView)
	}

	// 7. BrainActivityMsg updates facts list
	_ = brainRepo.SaveFact(ctx, domain.BrainFact{
		ID:      "fact-3",
		Topic:   "Decisión",
		Concept: "Ruta del Norte",
		Fact:    "El grupo decide cruzar por las montañas",
		Type:    domain.FactTypeDecision,
	})
	sidebar, actCmd := sidebar.Update(messages.BrainActivityMsg{
		Event: domain.BrainActivityEvent{
			Type:        "saved",
			FactCount:   1,
			Description: "🧠 Memorizado: 1 hecho",
		},
	})
	if actCmd != nil {
		msg := actCmd()
		sidebar, _ = sidebar.Update(msg)
	}
	if len(sidebar.BrainFacts) != 1 {
		t.Errorf("expected 1 fact in sidebar after BrainActivityMsg reload, got %d", len(sidebar.BrainFacts))
	}
}

func TestEditorGotoLine(t *testing.T) {
	styles := theme.DefaultStyles
	editor := components.NewEditorModel(styles)
	editor.SetSize(60, 20)

	content := "Línea 1\nLínea 2\nLínea 3\nLínea 4\nLínea 5"
	editor, _ = editor.Update(messages.ChapterSelectedMsg{
		Chapter: domain.Chapter{
			ID:      "chap-1",
			Title:   "Capítulo 1",
			Content: content,
		},
	})

	// GotoLine 3
	editor.GotoLine(3)
	// Line() in textarea is 0-indexed, so line 3 is index 2
	if editor.Line() != 2 {
		t.Errorf("expected line index 2 for line 3, got %d", editor.Line())
	}

	// GotoLine 1
	editor.GotoLine(1)
	if editor.Line() != 0 {
		t.Errorf("expected line index 0 for line 1, got %d", editor.Line())
	}

	// SetCursorPosition(4, 2)
	editor.SetCursorPosition(4, 2)
	if editor.Line() != 3 {
		t.Errorf("expected line index 3 for line 4, got %d", editor.Line())
	}
}

func TestSearchModalComponent(t *testing.T) {
	styles := theme.DefaultStyles
	modal := components.NewSearchModalModel(styles)
	modal.SetSize(100, 30)

	// Initially inactive
	if modal.Active {
		t.Errorf("expected search modal to be inactive initially")
	}
	if modal.View() != "" {
		t.Errorf("expected empty view when inactive")
	}

	// Open search modal via OpenGlobalSearchMsg
	modal, cmd := modal.Update(messages.OpenGlobalSearchMsg{})
	if !modal.Active {
		t.Errorf("expected search modal to be active after OpenGlobalSearchMsg")
	}
	if cmd == nil {
		t.Errorf("expected blink/focus command on open")
	}

	// Set matches
	matches := []domain.SearchMatch{
		{
			ChapterID:    "01_cap1",
			ChapterTitle: "Capítulo 1",
			LineNumber:   5,
			Column:       10,
			LineText:     "El gran dragón volaba sobre el valle.",
			MatchText:    "dragón",
		},
		{
			ChapterID:    "02_cap2",
			ChapterTitle: "Capítulo 2",
			LineNumber:   12,
			Column:       3,
			LineText:     "Un dragón dormía plácidamente.",
			MatchText:    "dragón",
		},
	}
	modal.SetMatches(matches)
	if len(modal.Matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(modal.Matches))
	}

	// View rendering contains match badge and chapter info
	view := modal.View()
	if !strings.Contains(view, "coincidencias") && !strings.Contains(view, "2") {
		t.Errorf("expected match summary in view: %s", view)
	}
	if !strings.Contains(view, "Capítulo 1") {
		t.Errorf("expected chapter 1 in view: %s", view)
	}

	// Test navigation with Down arrow
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyDown})
	if modal.SelectedMatch != 1 {
		t.Errorf("expected SelectedMatch 1 after KeyDown, got %d", modal.SelectedMatch)
	}

	// Test navigation with Up arrow
	modal, _ = modal.Update(tea.KeyMsg{Type: tea.KeyUp})
	if modal.SelectedMatch != 0 {
		t.Errorf("expected SelectedMatch 0 after KeyUp, got %d", modal.SelectedMatch)
	}

	// Test Enter to jump to match
	modal, cmd = modal.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if modal.Active {
		t.Errorf("expected modal to close on Enter jump")
	}
	if cmd == nil {
		t.Fatalf("expected JumpToMatchMsg command on Enter")
	}
	jumpMsg, ok := cmd().(messages.JumpToMatchMsg)
	if !ok {
		t.Fatalf("expected JumpToMatchMsg, got %T", cmd())
	}
	if jumpMsg.Match.ChapterID != "01_cap1" || jumpMsg.Match.LineNumber != 5 {
		t.Errorf("unexpected JumpToMatchMsg content: %+v", jumpMsg.Match)
	}

	// Reopen and test Tab focus switching and Replace input
	modal, _ = modal.Update(messages.OpenGlobalSearchMsg{})
	modal.SearchInput.SetValue("dragón")
	modal.ReplaceInput.SetValue("fénix")
	modal.IsReplaceMode = true
	modal.FocusIndex = 1 // focus replace input

	// Enter on replace input emits GlobalReplaceMsg
	modal, cmd = modal.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected GlobalReplaceMsg command on Enter from Replace input")
	}
	repMsg, ok := cmd().(messages.GlobalReplaceMsg)
	if !ok {
		t.Fatalf("expected GlobalReplaceMsg, got %T", cmd())
	}
	if repMsg.Query != "dragón" || repMsg.Replacement != "fénix" {
		t.Errorf("unexpected GlobalReplaceMsg: %+v", repMsg)
	}

	// Test Esc closes modal
	modal, cmd = modal.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if modal.Active {
		t.Errorf("expected modal to be inactive on Esc")
	}
	if cmd != nil {
		if _, ok := cmd().(messages.CloseGlobalSearchMsg); !ok {
			t.Errorf("expected CloseGlobalSearchMsg, got %T", cmd())
		}
	}

	// Test SearchFunc live search
	modal.SearchFunc = func(query string, caseSensitive bool) ([]domain.SearchMatch, error) {
		if query == "fénix" {
			return []domain.SearchMatch{
				{ChapterID: "chap-1", LineNumber: 1, MatchText: "fénix"},
			}, nil
		}
		return nil, nil
	}
	modal.SearchInput.SetValue("fénix")
	modal.PerformSearch()
	if len(modal.Matches) != 1 {
		t.Fatalf("expected 1 match from PerformSearch, got %d", len(modal.Matches))
	}
}

