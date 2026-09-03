package model_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/repository"
	"github.com/SalvucciFacundo/novel-tui/internal/service"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/model"
)

func TestRootModel_Init_NoPanicWhenNoInitialDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-init-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")
	configRepo := repository.NewFileConfigRepository(configPath)
	workspaceMgr := service.NewWorkspaceManager()

	root := model.NewRootModelWithConfig(configRepo, workspaceMgr, messages.ViewStateLauncher, "")

	// Init should not panic
	cmd := root.Init()
	if cmd != nil {
		_ = cmd()
	}
}

func TestRootModel_InitializationAndResizing(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-root-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	chapterRepo, err := repository.NewFileChapterRepository(tempDir)
	if err != nil {
		t.Fatalf("NewFileChapterRepository failed: %v", err)
	}
	characterRepo := repository.NewFileCharacterRepository(tempDir)

	root := model.NewRootModel(chapterRepo, characterRepo)

	// 1. Initial view before resize message
	initView := root.View()
	if !strings.Contains(initView, "Initializing Novel TUI") {
		t.Errorf("expected Initializing state, got %s", initView)
	}

	// 2. Small window warning
	smallSizeMsg := tea.WindowSizeMsg{Width: 40, Height: 10}
	updatedModel, _ := root.Update(smallSizeMsg)
	smallView := updatedModel.View()
	if !strings.Contains(smallView, "Terminal size too small") {
		t.Errorf("expected small window warning, got: %s", smallView)
	}

	// 3. Normal window layout in editor view
	normalSizeMsg := tea.WindowSizeMsg{Width: 100, Height: 30}
	updatedModel, _ = updatedModel.Update(normalSizeMsg)
	normalView := updatedModel.View()
	if strings.Contains(normalView, "Terminal size too small") {
		t.Errorf("expected normal view, but got warning: %s", normalView)
	}
	if !strings.Contains(normalView, "Inicio (Ctrl+H)") {
		t.Errorf("expected Navbar in editor view, got: %s", normalView)
	}
}

func TestRootModel_LauncherModeAndTransitions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-launcher-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")
	configRepo := repository.NewFileConfigRepository(configPath)
	workspaceMgr := service.NewWorkspaceManager()

	root := model.NewRootModelWithConfig(configRepo, workspaceMgr, messages.ViewStateLauncher, "")

	// Initialize window size
	m, _ := root.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	view := m.View()

	if !strings.Contains(view, "Acciones Rápidas") {
		t.Errorf("expected launcher in view: %s", view)
	}

	// Transition to LLM Config view
	m, _ = m.Update(messages.ChangeViewMsg{View: messages.ViewStateLLMConfig})
	llmView := m.View()
	if !strings.Contains(llmView, "Configuración de Inteligencia Artificial") {
		t.Errorf("expected LLM config view, got: %s", llmView)
	}

	// Transition back to Launcher
	m, _ = m.Update(messages.ChangeViewMsg{View: messages.ViewStateLauncher})
	launcherView := m.View()
	if !strings.Contains(launcherView, "Acciones Rápidas") {
		t.Errorf("expected Launcher view, got: %s", launcherView)
	}

	// Open Modal for New Novel
	m, _ = m.Update(messages.ShowModalMsg{
		Purpose: messages.ModalPurposeNewNovel,
		Title:   "Nueva Novela",
		Prompt:  "Título de la novela:",
	})
	modalView := m.View()
	if !strings.Contains(modalView, "Nueva Novela") {
		t.Errorf("expected modal view overlay, got: %s", modalView)
	}

	// Close Modal
	m, _ = m.Update(messages.HideModalMsg{})
	afterClose := m.View()
	if strings.Contains(afterClose, "Título de la novela:") {
		t.Errorf("expected modal to be hidden")
	}
}

func TestRootModel_CreateNovelAndTransitionToEditor(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-create-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")
	configRepo := repository.NewFileConfigRepository(configPath)
	cfg, _ := configRepo.Load()
	cfg.RootDir = tempDir
	_ = configRepo.Save(cfg)

	workspaceMgr := service.NewWorkspaceManager()

	root := model.NewRootModelWithConfig(configRepo, workspaceMgr, messages.ViewStateLauncher, "")
	m, _ := root.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Dispatch CreateNovelMsg
	m, cmd := m.Update(messages.CreateNovelMsg{Title: "Las Crónicas de Noria"})
	if cmd == nil {
		t.Fatalf("expected command after CreateNovelMsg")
	}
	msg := cmd()
	openMsg, ok := msg.(messages.OpenNovelMsg)
	if !ok {
		t.Fatalf("expected OpenNovelMsg, got: %+v", msg)
	}

	// Update with OpenNovelMsg
	m, _ = m.Update(openMsg)

	// Verify view is now in Editor mode with Navbar
	editorView := m.View()
	if strings.Contains(editorView, "Acciones Rápidas") {
		t.Errorf("expected Editor view, but still on Launcher")
	}
	if !strings.Contains(editorView, "Las Crónicas de Noria") {
		t.Errorf("expected novel title in Navbar breadcrumbs, got:\n%s", editorView)
	}
}

func TestRootModel_MouseRouting(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "novel-tui-root-mouse-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")
	configRepo := repository.NewFileConfigRepository(configPath)
	workspaceMgr := service.NewWorkspaceManager()

	root := model.NewRootModelWithConfig(configRepo, workspaceMgr, messages.ViewStateLauncher, "")
	m, _ := root.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// 1. Mouse click in Launcher routes to launcher
	m, cmd := m.Update(tea.MouseMsg{
		X:    15,
		Y:    19,
		Type: tea.MouseLeft,
	})
	if cmd != nil {
		msg := cmd()
		if viewMsg, ok := msg.(messages.ChangeViewMsg); ok {
			m, _ = m.Update(viewMsg)
		}
	}
	view := m.View()
	if !strings.Contains(view, "Configuración de Inteligencia Artificial") {
		t.Errorf("expected LLM config view after mouse click, got: %s", view)
	}

	// 2. Switch to Editor view and test Navbar click routing (Y == 0)
	m, _ = m.Update(messages.ChangeViewMsg{View: messages.ViewStateEditor})
	// Click Inicio pill at Y = 0, X = 5
	m, cmd = m.Update(tea.MouseMsg{
		X:    5,
		Y:    0,
		Type: tea.MouseLeft,
	})
	if cmd != nil {
		msg := cmd()
		if viewMsg, ok := msg.(messages.ChangeViewMsg); ok {
			m, _ = m.Update(viewMsg)
		}
	}
	view = m.View()
	if !strings.Contains(view, "Acciones Rápidas") {
		t.Errorf("expected Launcher view after clicking Inicio pill in Navbar, got: %s", view)
	}
}

func TestRootModel_KeybindingIsolation_CtrlN(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	configRepo := repository.NewFileConfigRepository(configPath)
	workspaceMgr := service.NewWorkspaceManager()

	root := model.NewRootModelWithConfig(configRepo, workspaceMgr, messages.ViewStateEditor, tempDir)
	m, _ := root.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// 1. Focus on Editor -> Ctrl+N MUST open New Chapter modal
	m, _ = m.Update(messages.FocusMsg{Target: messages.FocusEditor})
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	if cmd == nil {
		t.Fatalf("expected command on Ctrl+N when focused on Editor")
	}
	msg := cmd()
	showMsg, ok := msg.(messages.ShowModalMsg)
	if !ok || showMsg.Purpose != messages.ModalPurposeNewChapter {
		t.Errorf("expected ShowModalMsg(NewChapter), got: %+v", msg)
	}

	// Close modal
	m, _ = m.Update(messages.HideModalMsg{})

	// 2. Focus on Chat Drawer -> Ctrl+N MUST NOT open New Chapter modal
	m, _ = m.Update(messages.ToggleChatDrawerMsg{})
	m, _ = m.Update(messages.FocusMsg{Target: messages.FocusChat})
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	if cmd != nil {
		msg := cmd()
		if showMsg, ok := msg.(messages.ShowModalMsg); ok && showMsg.Purpose == messages.ModalPurposeNewChapter {
			t.Errorf("Ctrl+N when focused on Chat should NOT trigger NewChapter modal")
		}
	}
}

func TestRootModel_Global_CtrlH(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	configRepo := repository.NewFileConfigRepository(configPath)
	workspaceMgr := service.NewWorkspaceManager()

	root := model.NewRootModelWithConfig(configRepo, workspaceMgr, messages.ViewStateEditor, tempDir)
	m, _ := root.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Press Ctrl+H anywhere in editor view -> return to Launcher
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	if cmd == nil {
		t.Fatalf("expected command on Ctrl+H")
	}
	msg := cmd()
	viewMsg, ok := msg.(messages.ChangeViewMsg)
	if !ok || viewMsg.View != messages.ViewStateLauncher {
		t.Errorf("expected ChangeViewMsg(Launcher) on Ctrl+H, got: %+v", msg)
	}
}

func TestRootModel_ChatDrawer_ToggleAndNavigation(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	configRepo := repository.NewFileConfigRepository(configPath)
	workspaceMgr := service.NewWorkspaceManager()

	root := model.NewRootModelWithConfig(configRepo, workspaceMgr, messages.ViewStateEditor, tempDir)
	m, _ := root.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// 1. Initial view without drawer
	initialView := m.View()
	if strings.Contains(initialView, "Inicia una conversación") || strings.Contains(initialView, "[Medio]") {
		t.Errorf("expected chat drawer to be hidden initially")
	}

	// 2. Press Ctrl+A to toggle chat drawer open
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	if cmd != nil {
		msg := cmd()
		m, _ = m.Update(msg)
	}
	drawerView := m.View()
	if !strings.Contains(drawerView, "[Medio]") {
		t.Errorf("expected chat drawer to be visible after Ctrl+A, got:\n%s", drawerView)
	}

	// 3. Tab cycling with open drawer: FocusChat -> FocusSidebar -> FocusEditor -> FocusChat
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	// 4. Toggle chat drawer closed with Ctrl+A
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	if cmd != nil {
		msg := cmd()
		m, _ = m.Update(msg)
	}
	closedView := m.View()
	if strings.Contains(closedView, "[Medio]") {
		t.Errorf("expected chat drawer to be hidden after second Ctrl+A")
	}
}

func TestRootModel_ChatDrawer_TokenStreaming(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	configRepo := repository.NewFileConfigRepository(configPath)
	workspaceMgr := service.NewWorkspaceManager()

	root := model.NewRootModelWithConfig(configRepo, workspaceMgr, messages.ViewStateEditor, tempDir)
	m, _ := root.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m, _ = m.Update(messages.ToggleChatDrawerMsg{})

	// Receive streaming tokens
	m, _ = m.Update(messages.TokenReceivedMsg{Content: "Respuesta"})
	m, _ = m.Update(messages.TokenReceivedMsg{Content: " generada"})
	m, _ = m.Update(messages.StreamFinishedMsg{})

	view := m.View()
	if !strings.Contains(view, "Respuesta generada") {
		t.Errorf("expected streamed response in chat drawer view, got:\n%s", view)
	}
}

func TestRootModel_ChatDrawer_BrainActivity(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	configRepo := repository.NewFileConfigRepository(configPath)
	workspaceMgr := service.NewWorkspaceManager()

	root := model.NewRootModelWithConfig(configRepo, workspaceMgr, messages.ViewStateEditor, tempDir)
	m, _ := root.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m, _ = m.Update(messages.ToggleChatDrawerMsg{})

	// Send BrainActivityMsg
	m, _ = m.Update(messages.BrainActivityMsg{
		Event: domain.BrainActivityEvent{
			Type:        "saved",
			FactCount:   2,
			Description: "🧠 [Brain] Memorizado: 2 hecho(s) (Kuno, Espada)",
		},
	})

	view := m.View()
	if !strings.Contains(view, "🧠 [Brain] Memorizado") {
		t.Errorf("expected brain notification in chat view, got:\n%s", view)
	}
}

func TestRootModel_TimelineIntegration(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	configRepo := repository.NewFileConfigRepository(configPath)
	workspaceMgr := service.NewWorkspaceManager()

	novelMeta, err := workspaceMgr.CreateNovel(tempDir, "Novela con Timeline")
	if err != nil {
		t.Fatalf("failed to create novel: %v", err)
	}

	// Add timeline events to .novel/brain.db
	brainDbPath := filepath.Join(novelMeta.AbsolutePath, ".novel", "brain.db")
	brainRepo, err := repository.NewSQLiteBrainRepository(brainDbPath)
	if err != nil {
		t.Fatalf("failed to open brain db: %v", err)
	}
	_ = brainRepo.SaveTimelineEvent(context.Background(), domain.TimelineEvent{
		ID:                 "tl-1",
		ChronologicalOrder: 1,
		Period:             "Era Antigua",
		Title:              "Gran Cataclismo",
		Description:        "El mundo se dividió en dos continentes",
		Characters:         []string{"Ancestros"},
	})
	_ = brainRepo.Close()

	root := model.NewRootModelWithConfig(configRepo, workspaceMgr, messages.ViewStateLauncher, tempDir)
	m, _ := root.Update(tea.WindowSizeMsg{Width: 140, Height: 40})

	// Open novel
	m, openCmd := m.Update(messages.OpenNovelMsg{Path: novelMeta.AbsolutePath})
	if openCmd != nil {
		msg := openCmd()
		if msg != nil {
			m, _ = m.Update(msg)
		}
	}

	// Switch sidebar tab to TabBrain (tab 3)
	m, _ = m.Update(messages.SelectSidebarTabMsg{Tab: 3})
	m, _ = m.Update(messages.FocusMsg{Target: messages.FocusSidebar})

	// Trigger BrainActivityMsg reload
	m, actCmd := m.Update(messages.BrainActivityMsg{
		Event: domain.BrainActivityEvent{
			Type:        "saved",
			FactCount:   1,
			Description: "🧠 [Brain] Memorizado",
		},
	})
	if actCmd != nil {
		msg := actCmd()
		if msg != nil {
			m, _ = m.Update(msg)
		}
	}

	// Toggle to Timeline view with 't'
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})

	view := m.View()
	if !strings.Contains(view, "Gran Cataclismo") {
		t.Errorf("expected 'Gran Cataclismo' timeline event in editor view sidebar: %s", view)
	}
	if !strings.Contains(view, "Era Antigua") {
		t.Errorf("expected 'Era Antigua' period header in sidebar timeline view: %s", view)
	}
}

func TestRootModel_OpenNovel_SetsBrainRepoInSidebar(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	configRepo := repository.NewFileConfigRepository(configPath)
	workspaceMgr := service.NewWorkspaceManager()

	novelMeta, err := workspaceMgr.CreateNovel(tempDir, "Novela con Brain")
	if err != nil {
		t.Fatalf("failed to create novel: %v", err)
	}

	// Add facts to .novel/brain.db
	brainDbPath := filepath.Join(novelMeta.AbsolutePath, ".novel", "brain.db")
	brainRepo, err := repository.NewSQLiteBrainRepository(brainDbPath)
	if err != nil {
		t.Fatalf("failed to open brain db: %v", err)
	}
	_ = brainRepo.SaveFact(context.Background(), domain.BrainFact{
		ID:      "bf-1",
		Topic:   "Lore",
		Concept: "Torre Oscura",
		Fact:    "Una torre de piedra negra en el horizonte",
		Type:    domain.FactTypeLore,
	})
	_ = brainRepo.Close()

	root := model.NewRootModelWithConfig(configRepo, workspaceMgr, messages.ViewStateLauncher, tempDir)
	m, _ := root.Update(tea.WindowSizeMsg{Width: 140, Height: 40})

	// Open novel
	m, openCmd := m.Update(messages.OpenNovelMsg{Path: novelMeta.AbsolutePath})
	if openCmd != nil {
		msg := openCmd()
		if msg != nil {
			m, _ = m.Update(msg)
		}
	}

	// Switch sidebar tab to TabBrain (tab 3)
	m, _ = m.Update(messages.SelectSidebarTabMsg{Tab: 3})
	// Trigger brain activity to reload
	m, actCmd := m.Update(messages.BrainActivityMsg{
		Event: domain.BrainActivityEvent{
			Type:        "saved",
			FactCount:   1,
			Description: "🧠 [Brain] Memorizado",
		},
	})
	if actCmd != nil {
		msg := actCmd()
		if msg != nil {
			m, _ = m.Update(msg)
		}
	}

	view := m.View()
	if !strings.Contains(view, "Torre Oscura") {
		t.Errorf("expected Torre Oscura fact to be visible in editor view sidebar: %s", view)
	}
}

func TestRootModel_GlobalSearchAndJumpToMatch(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	configRepo := repository.NewFileConfigRepository(configPath)
	workspaceMgr := service.NewWorkspaceManager()

	novelMeta, err := workspaceMgr.CreateNovel(tempDir, "Novela de Busqueda")
	if err != nil {
		t.Fatalf("failed to create novel: %v", err)
	}

	// Create chapters
	chapRepo, err := repository.NewFileChapterRepository(novelMeta.AbsolutePath)
	if err != nil {
		t.Fatalf("failed to init chap repo: %v", err)
	}
	_ = chapRepo.SaveContent("01_capitulo_1", "El dragón descansaba en la caverna.")
	ch2, _ := chapRepo.Create("Capítulo 2")
	_ = chapRepo.SaveContent(ch2.ID, "Línea 1 vacía\nEl dragón rugió en las sombras.")

	root := model.NewRootModelWithConfig(configRepo, workspaceMgr, messages.ViewStateLauncher, tempDir)
	m, _ := root.Update(tea.WindowSizeMsg{Width: 140, Height: 40})

	// Open novel
	m, _ = m.Update(messages.OpenNovelMsg{Path: novelMeta.AbsolutePath})

	// Trigger Search via Ctrl+F
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
	if cmd != nil {
		msg := cmd()
		if msg != nil {
			m, _ = m.Update(msg)
		}
	}

	view := m.View()
	if !strings.Contains(view, "Búsqueda") {
		t.Errorf("expected search modal to be rendered in view, got:\n%s", view)
	}

	// Send JumpToMatchMsg targeting Chapter 2, line 2
	match := domain.SearchMatch{
		ChapterID:    ch2.ID,
		ChapterTitle: ch2.Title,
		FilePath:     ch2.FilePath,
		LineNumber:   2,
		Column:       4,
		LineText:     "El dragón rugió en las sombras.",
		MatchText:    "dragón",
	}
	m, _ = m.Update(messages.JumpToMatchMsg{Match: match})

	postJumpView := m.View()
	if strings.Contains(postJumpView, "Búsqueda Global") {
		t.Errorf("expected search modal to close after JumpToMatch")
	}
	if !strings.Contains(postJumpView, "Capítulo 2") {
		t.Errorf("expected navbar / editor to show Capítulo 2 after jump, got:\n%s", postJumpView)
	}
}

func TestRootModel_TimelineInChatPrompt(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	configRepo := repository.NewFileConfigRepository(configPath)
	workspaceMgr := service.NewWorkspaceManager()

	novelMeta, err := workspaceMgr.CreateNovel(tempDir, "Novela Prompt Test")
	if err != nil {
		t.Fatalf("failed to create novel: %v", err)
	}

	brainDbPath := filepath.Join(novelMeta.AbsolutePath, ".novel", "brain.db")
	brainRepo, err := repository.NewSQLiteBrainRepository(brainDbPath)
	if err != nil {
		t.Fatalf("failed to open brain db: %v", err)
	}
	_ = brainRepo.SaveTimelineEvent(context.Background(), domain.TimelineEvent{
		ID:                 "tl-prompt-1",
		ChronologicalOrder: 1,
		Period:             "Prólogo",
		Title:              "Despertar del Héroe",
		Description:        "El protagonista despierta sin recuerdos",
	})
	_ = brainRepo.Close()

	root := model.NewRootModelWithConfig(configRepo, workspaceMgr, messages.ViewStateLauncher, tempDir)
	m, _ := root.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	m, _ = m.Update(messages.OpenNovelMsg{Path: novelMeta.AbsolutePath})

	// Dispatch SendChatMessageMsg
	_, cmd := m.Update(messages.SendChatMessageMsg{
		Content:     "¿Qué ocurrió en el prólogo?",
		EffortLevel: domain.EffortMedium,
	})

	// Command should be created (even if provider is offline, cmd is generated)
	if cmd == nil {
		t.Fatalf("expected stream command on SendChatMessageMsg")
	}
}

func TestRootModel_GlobalReplace(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	configRepo := repository.NewFileConfigRepository(configPath)
	workspaceMgr := service.NewWorkspaceManager()

	novelMeta, err := workspaceMgr.CreateNovel(tempDir, "Novela de Reemplazo")
	if err != nil {
		t.Fatalf("failed to create novel: %v", err)
	}

	chapRepo, err := repository.NewFileChapterRepository(novelMeta.AbsolutePath)
	if err != nil {
		t.Fatalf("failed to init chap repo: %v", err)
	}
	_ = chapRepo.SaveContent("01_capitulo_1", "El dragón dormía.")
	ch2, _ := chapRepo.Create("Capítulo 2")
	_ = chapRepo.SaveContent(ch2.ID, "Un dragón despierto.")

	root := model.NewRootModelWithConfig(configRepo, workspaceMgr, messages.ViewStateLauncher, tempDir)
	m, _ := root.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	m, _ = m.Update(messages.OpenNovelMsg{Path: novelMeta.AbsolutePath})

	// Execute Global Replace
	m, cmd := m.Update(messages.GlobalReplaceMsg{
		Query:         "dragón",
		Replacement:   "fénix",
		CaseSensitive: false,
	})

	if cmd != nil {
		msg := cmd()
		if msg != nil {
			m, _ = m.Update(msg)
		}
	}

	// Verify files updated
	c1Content, _ := chapRepo.LoadContent("01_capitulo_1")
	if !strings.Contains(c1Content, "fénix") {
		t.Errorf("expected c1 content to contain fénix, got: %s", c1Content)
	}
	c2Content, _ := chapRepo.LoadContent(ch2.ID)
	if !strings.Contains(c2Content, "fénix") {
		t.Errorf("expected c2 content to contain fénix, got: %s", c2Content)
	}
}

